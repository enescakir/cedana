import csv
import os
import shutil
import subprocess
import time
import requests 
import profile_pb2
from google.cloud import bigquery
from google.cloud.bigquery import LoadJobConfig, SourceFormat
import json


import psutil

benchmarking_dir = "benchmarks"
output_dir = "benchmark_results"


def setup():
    # download benchmarking repo
    repo_url="https://github.com/cedana/cedana-benchmarks"
    subprocess.run(["git", "clone", repo_url, benchmarking_dir])

    # make folder for storing results
    os.makedirs(output_dir, exist_ok=True)

    # get cedana daemon pid from pid file 
    with open("/run/cedana.pid", "r") as file:
        daemon_pid = int(file.read().strip())

    print("found daemon running at pid {}".format(daemon_pid))

    return daemon_pid

def cleanup():
    shutil.rmtree(benchmarking_dir)

def get_process_by_pid(pid):
    # Get a psutil.Process object for the given pid
    try:
        return psutil.Process(pid)
    except psutil.NoSuchProcess:
        print("No process found with PID", pid)
        return None

def measure_disk_usage(checkpoint_dir):
    return sum(os.path.getsize(os.path.join(dirpath, filename)) for dirpath, dirnames, filenames in os.walk(checkpoint_dir) for filename in filenames)


def start_pprof(filename): 
    pprof_base_url = "http://localhost:6060"
    resp = requests.get(f"{pprof_base_url}/start-profiling?prefix={filename}")
    if resp.status_code != 200:
        print("error from profiler: {}".format(resp.text))

def stop_pprof(filename):
    pprof_base_url = "http://localhost:6060"
    resp = requests.get(f"{pprof_base_url}/stop-profiling?filename={filename}")
    if resp.status_code != 200:
        print("error from profiler: {}".format(resp.text))


def start_recording(pid):
    initial_data = {}
    try:
        p = psutil.Process(pid)
        initial_data['cpu_times'] = p.cpu_times()
        initial_data['memory'] = p.memory_full_info().uss
        initial_data['disk_io'] = psutil.disk_io_counters()
        initial_data['cpu_percent'] = p.cpu_percent(interval=None)
        initial_data['time'] = psutil.cpu_times()
    except psutil.NoSuchProcess:
        print(f"No such process with PID {pid}")

    return initial_data

def stop_recording(operation_type, pid, initial_data, jobID, completed_at, started_at, process_stats):
    p = psutil.Process(pid)
    current_cpu_times = p.cpu_times()
    cpu_time_user_diff = current_cpu_times.user - initial_data['cpu_times'].user
    cpu_time_system_diff = current_cpu_times.system - initial_data['cpu_times'].system
    cpu_utilization = (cpu_time_user_diff + cpu_time_system_diff)
    
    cpu_time_total_diff = (cpu_time_user_diff + cpu_time_system_diff)
         
    #Calculate the total time all CPUs have spent since we started recording
    current_time = psutil.cpu_times()
    cpu_total_time_diff = sum(getattr(current_time, attr) - getattr(initial_data['time'], attr)
                                  for attr in ['user', 'system', 'idle'])

    # Calculate the percentage of CPU utilization
    cpu_percent = 100 * cpu_time_total_diff / cpu_total_time_diff if cpu_total_time_diff else 0

    # Memory usage in KB
    current_memory = p.memory_full_info().uss
    memory_used_kb = (current_memory - initial_data['memory']) / 1024
        
    # Disk I/O
    current_disk_io = psutil.disk_io_counters()
    read_count_diff = current_disk_io.read_count - initial_data['disk_io'].read_count
    write_count_diff = current_disk_io.write_count - initial_data['disk_io'].write_count
    read_bytes_diff = current_disk_io.read_bytes - initial_data['disk_io'].read_bytes
    write_bytes_diff = current_disk_io.write_bytes - initial_data['disk_io'].write_bytes

    # read from profiling json 
    network_op = ""
    compress_op = ""
    if operation_type == "checkpoint":
        network_op = "upload"
        compress_op = "compress"
    else: 
        network_op = "download"
        compress_op = "decompress"

    with open("/var/log/cedana-profile.json", 'r') as f:
        profiling_data = json.load(f)
        op_duration = profiling_data[operation_type]
        network_duration = profiling_data[network_op]
        compress_duration = profiling_data[compress_op]

    with open("benchmark_output.csv", mode='a', newline='') as file:
        writer = csv.writer(file)
        # Write the headers if the file is new
        if file.tell() == 0:
            writer.writerow([
                'Timestamp', 
                'Job ID', 
                'Operation Type',
                'Memory Used Target (KB)',
                'Memory Used Daemon', 
                'Write Count', 
                'Read Count', 
                'Write (MB)', 
                'Read Bytes (MB)', 
                'CPU Utilization (Secs)', 
                'CPU Used (Percent)', 
                'Total Duration',
                'Operation Duration',
                'Network Duration',
                "Compression Duration",
                ])
        
        # Write the resource usage data
        writer.writerow([
            time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(time.time())),
            jobID,
            operation_type,
            process_stats['memory_kb'],
            memory_used_kb,
            write_count_diff,
            read_count_diff,
            write_bytes_diff / (1024 * 1024), # convert to MB
            read_bytes_diff / (1024 * 1024), # convert to MB
            cpu_utilization,
            cpu_percent,
            completed_at - started_at,
            op_duration,
            network_duration,
            compress_duration
        ])

        # delete profile file after
        os.remove("/var/log/cedana-profile.json")

def analyze_pprof(filename):
    pass 

def run_checkpoint(daemonPID, jobID, iteration, output_dir, process_stats): 
    chkpt_cmd = "sudo -E ./cedana dump job {} -d tmp".format(jobID+"-"+str(iteration))

    # initial data here is fine - we want to measure impact of daemon on system 
    initial_data = start_recording(daemonPID)
    cpu_profile_filename = "{}/cpu_{}_{}_checkpoint".format(output_dir, jobID, iteration)
    start_pprof(cpu_profile_filename)

    checkpoint_started_at = time.monotonic_ns() 
    p = subprocess.Popen(["sh", "-c", chkpt_cmd], stdout=subprocess.PIPE)
    # used for capturing full time instead of directly exiting
    p.wait()

    # these values have an error range in 35ms! I blame Python?
    checkpoint_completed_at = time.monotonic_ns()  
    stop_recording("checkpoint", daemonPID, initial_data, jobID, checkpoint_completed_at, checkpoint_started_at, process_stats)
    stop_pprof(cpu_profile_filename)

def run_restore(daemonPID, jobID, iteration, output_dir):
    restore_cmd = "sudo -E ./cedana restore job {}".format(jobID+"-"+str(iteration))

    initial_data = start_recording(daemonPID)
    cpu_profile_filename = "{}/cpu_{}_{}_restore".format(output_dir, jobID, iteration)
    start_pprof(cpu_profile_filename)

    restore_started_at = time.monotonic_ns() 
    p =subprocess.Popen(["sh", "-c", restore_cmd], stdout=subprocess.PIPE)
    p.wait()

    restore_completed_at = time.monotonic_ns()

    # nil value here
    process_stats = {}
    process_stats["memory_kb"] = 0
    stop_recording("restore", daemonPID, initial_data, jobID, restore_completed_at, restore_started_at, process_stats)
    stop_pprof(cpu_profile_filename)

def run_exec(cmd, jobID): 
    process_stats = {}
    exec_cmd = "./cedana exec {} {}".format(cmd, jobID)

    process = subprocess.Popen(["sh", "-c", exec_cmd], stdout=subprocess.PIPE)
    pid = int(process.communicate()[0].decode().strip())
    process_stats['pid'] = pid

    psutil_process = psutil.Process(pid)
    process_stats['memory_kb'] = psutil_process.memory_full_info().uss / 1024 # convert to KB

    return process_stats 

def push_to_bigquery():
    client = bigquery.Client()

    dataset_id = 'devtest'
    table_id = 'benchmarks'

    csv_file_path = 'benchmark_output.csv'

    job_config = LoadJobConfig(
        source_format=SourceFormat.CSV,
        skip_leading_rows=1,  # Change this according to your CSV file
        autodetect=True,  # Auto-detect schema if the table doesn't exist
        write_disposition="WRITE_APPEND",  # Options are WRITE_APPEND, WRITE_EMPTY, WRITE_TRUNCATE
)

    dataset_ref = client.dataset(dataset_id)
    table_ref = dataset_ref.table(table_id)

    # API request to start the job
    with open(csv_file_path, "rb") as source_file:
        load_job = client.load_table_from_file(
            source_file,
            table_ref,
            job_config=job_config
        )  

    load_job.result()  

    if load_job.errors is not None:
        print('Errors:', load_job.errors)
    else:
        print('Job finished successfully.')

    # Get the table details
    table = client.get_table(table_ref)  
    print('Loaded {} rows to {}'.format(table.num_rows, table_id))
  
def main(): 
    daemon_pid = setup()
    jobIDs = [
        "server",
        "loop",
        "regression",
        "nn-1gb"
    ]
    cmds = [
        "./benchmarks/server",
        "./benchmarks/test.sh",
        "'python3 benchmarks/regression/main.py'",
        "'python3 benchmarks/1gb_pytorch.py'"
    ]

    # run in a loop 
    num_samples = 10 
    for x in range(len(jobIDs)): 
        print("Starting benchmarks for job {} with command {}".format(jobIDs[x], cmds[x]))
        jobID = jobIDs[x]
        for y in range(num_samples):
            # we pass a job ID + iteration to generate a unique one every time. 
            # sometimes in docker containers, the db file doesn't update fast (especially for the quick benchmarks) and 
            # we end up getting a killed PID.
            process_stats = run_exec(cmds[x], jobID+"-"+str(y))
            # wait a few seconds for memory to allocate 
            time.sleep(5)

            # we don't mutate jobID for checkpoint/restore here so we can pass the unadulterated one to our csv  
            run_checkpoint(daemon_pid, jobID, y, output_dir, process_stats)
            time.sleep(1)

            run_restore(daemon_pid, jobID, y, output_dir)
            time.sleep(1)

    push_to_bigquery()

    # delete benchmarking folder
    cleanup()

main()
