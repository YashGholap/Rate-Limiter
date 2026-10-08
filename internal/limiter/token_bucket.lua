local key = KEYS[1]

local capacity = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])

local time = redis.call("TIME")
local now_seconds = time[1] + (time[2] / 1000000)

local tokens = redis.call("HGET", key, "tokens")
local last_refill = redis.call("HGET", key, "last_refill")

if tokens == false then
    tokens = capacity
    last_refill = now_seconds
else
    tokens = tonumber(tokens)
    last_refill = tonumber(last_refill)
end

local elapsed = now_seconds - last_refill
local tokens_to_add = elapsed * refill_rate

tokens = math.min(capacity, tokens + tokens_to_add)
last_refill = now_seconds

if tokens < 1 then
    redis.call(
        "HSET",
        key,
        "tokens", tokens,
        "last_refill", last_refill
    )

    return 0
end

tokens = tokens - 1

redis.call(
    "HSET",
    key,
    "tokens", tokens,
    "last_refill", last_refill
)

return 1