import httpx

from stress import cli


class FakeClock:
    """Stands in for time.monotonic and time.sleep, so no test really waits."""

    def __init__(self):
        self.now = 0.0

    def __call__(self):
        return self.now

    def sleep(self, seconds):
        self.now += seconds


def client_for(handler):
    return httpx.Client(transport=httpx.MockTransport(handler))


def wait(client, urls, clock, timeout=10):
    return cli.wait_until_healthy(
        client, urls, timeout=timeout, interval=2, sleep=clock.sleep, clock=clock
    )


def test_retries_until_a_server_starts_listening():
    attempts = []

    def handler(request):
        attempts.append(request.url)
        if len(attempts) < 3:
            raise httpx.ConnectError("connection refused", request=request)
        return httpx.Response(200)

    clock = FakeClock()
    codes = wait(client_for(handler), ["http://core/api/hello"], clock)
    assert codes == {"http://core/api/hello": 200}
    assert len(attempts) == 3
    assert clock.now == 4


def test_gives_up_at_the_deadline_with_the_last_status():
    clock = FakeClock()
    codes = wait(client_for(lambda request: httpx.Response(503)), ["http://core/api/hello"], clock)
    assert codes == {"http://core/api/hello": 503}
    assert clock.now == 10


def test_the_deadline_is_shared_across_urls():
    def handler(request):
        if request.url.host == "down":
            raise httpx.ConnectError("connection refused", request=request)
        return httpx.Response(200)

    clock = FakeClock()
    codes = wait(client_for(handler), ["http://down/a", "http://down/b", "http://up/c"], clock)
    assert codes == {"http://down/a": 0, "http://down/b": 0, "http://up/c": 200}
    assert clock.now == 10


def test_parse_args_preflight_timeout():
    assert cli.parse_args([]).preflight_timeout == cli.PREFLIGHT_TIMEOUT_S
    assert cli.parse_args(["--preflight-timeout", "30"]).preflight_timeout == 30
