#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

struct kill_event {
    __u64 timestamp_ns;
    __u32 sender_pid;
    __u32 target_pid;
    __u32 signal;
};

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 20);
} events SEC(".maps");

SEC("tracepoint/syscalls/sys_enter_kill")
int trace_kill(struct trace_event_raw_sys_enter *ctx)
{
    struct kill_event *event;

    event = bpf_ringbuf_reserve(&events, sizeof(*event), 0);
    if (!event)
        return 0;

    event->timestamp_ns = bpf_ktime_get_ns();
    event->sender_pid = bpf_get_current_pid_tgid() >> 32;
    event->target_pid = (__u32)ctx->args[0];
    event->signal = (__u32)ctx->args[1];
    bpf_printk("SO1_PR2_KILL sender=%d target=%d signal=%d", event->sender_pid, event->target_pid, event->signal);
    bpf_ringbuf_submit(event, 0);
    return 0;
}

char LICENSE[] SEC("license") = "GPL";
