namespace go base (cpp.map = "aaa")

cpp_include "unordered_map.h"
cpp_include "unordered_set.h"
cpp_include "string"
cpp_include "vector"

include "a.thrift"

enum StatusCode {
    Unknown = 0,
    Success = 1,
}

const string ServiceName = "demo"
const list<string> SupportedApps = ["douyin", "kuaishou", "Bilibili"]

typedef i32 AppVersion

struct PingRequest {
    1: required string Name,
    2: optional string company,
}

struct PingResponse {
    1: optional string Message,
}

/*
    DemoService says Hi to every request.
    Hello world!
*/
service DemoService {
    // Ping says hello to you.
    PingResponse Ping (1: PingRequest req) (api.get = "/v1/ping"),
}
