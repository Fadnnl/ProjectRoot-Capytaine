import grpc
import worker.testworker_pb2 as testworker_pb2
import worker.testworker_pb2_grpc as testworker_pb2_grpc

def main():
    # connect ke worker container (karena kita expose ke host di 50051)
    channel = grpc.insecure_channel("localhost:50051")
    stub = testworker_pb2_grpc.TestWorkerStub(channel)

    # bikin request
    request = testworker_pb2.SimRequest(
        scenario="demo_scenario",
        iterations=5,
        concurrency=2,
        params={"foo": "bar"}
    )

    # call gRPC method
    response = stub.RunSimulation(request)
    print("Response:")
    print("  status          :", response.status)
    print("  message         :", response.message)
    print("  elapsed_seconds :", response.elapsed_seconds)
    print("  iterations_done :", response.iterations_done)

if __name__ == "__main__":
    main()
