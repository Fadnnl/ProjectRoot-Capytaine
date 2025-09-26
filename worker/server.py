# worker/server.py
import time
import concurrent.futures
from concurrent.futures import ThreadPoolExecutor, as_completed
import grpc
import testworker_pb2, testworker_pb2_grpc
import capytaine as cpt
from capytaine.meshes.predefined import mesh_sphere
import os
from statsd import StatsClient
import math

# statsd client (graphite/statsd)
STATSD_HOST = os.getenv("STATSD_HOST", "graphite")
STATSD_PORT = int(os.getenv("STATSD_PORT", "8125"))
stats = StatsClient(STATSD_HOST, STATSD_PORT, prefix="capytest")

def run_single_sim(i, params):
    t0 = time.time()
    try:
        # Contoh simulasi kecil dengan Capytaine
       
        sphere = mesh_sphere(radius=1.0, nlat=8, nlong=8)
    except TypeError:
        sphere = mesh_sphere(radius=1.0, ntheta=8, nphi=8)


        # Pakai omega (frekuensi sudut), bukan wave_period
        wave_period = 5.0
        omega = 2 * math.pi / wave_period

        problem = cpt.DiffractionProblem(
            body=sphere,
            wave_direction=0.0,
            omega=omega
        )

        solver = cpt.BEMSolver()
        res = solver.solve(problem)

        elapsed = (time.time() - t0) * 1000.0
        stats.timing("iteration.latency_ms", elapsed)
        stats.incr("iteration.success")
        return True, elapsed
    except Exception as e:
        print(f"[ERROR] Iteration {i} failed: {repr(e)}", flush=True)
        stats.incr("iteration.error")
        return False, str(e)



class TestWorkerServicer(testworker_pb2_grpc.TestWorkerServicer):
    def RunSimulation(self, request, context):
        iterations = int(request.iterations or 100)
        concurrency = int(request.concurrency or 8)
        start = time.time()

        successes = 0
        failures = 0

        with ThreadPoolExecutor(max_workers=concurrency) as ex:
            futures = [ex.submit(run_single_sim, i, request.params) for i in range(iterations)]
            for f in concurrent.futures.as_completed(futures):
                ok, info = f.result()
                if ok:
                    successes += 1
                else:
                    failures += 1

        elapsed = time.time() - start
        stats.gauge("run.elapsed_seconds", elapsed)
        stats.gauge("run.iterations", iterations)
        stats.gauge("run.successes", successes)
        stats.gauge("run.failures", failures)

        return testworker_pb2.SimReply(
            status="ok" if failures == 0 else "partial",
            message=f"done (success={successes} fail={failures})",
            elapsed_seconds=elapsed,
            iterations_done=successes + failures
        )

def serve():
    server = grpc.server(ThreadPoolExecutor(max_workers=16))
    testworker_pb2_grpc.add_TestWorkerServicer_to_server(TestWorkerServicer(), server)
    server.add_insecure_port("[::]:50051")
    server.start()
    print("gRPC TestWorker serving on 0.0.0.0:50051")
    try:
        server.wait_for_termination()
    except KeyboardInterrupt:
        server.stop(0)

if __name__ == "__main__":
    serve()
