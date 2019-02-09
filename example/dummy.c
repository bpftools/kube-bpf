#include "elf/include/bpf.h"
#include "elf/include/bpf_map.h"
#include "elf/include/trace_printk.h"

#define SEC(NAME) __attribute__((section(NAME), used))

struct bpf_map_def SEC("maps/dummy_hash") dummy_hash = {
	.type = BPF_MAP_TYPE_HASH,
	.key_size = sizeof(int),
	.value_size = sizeof(unsigned int),
	.max_entries = 128,
};

SEC("tracepoint/raw_syscalls/sys_enter")
int tracepoint__raw__sys_enter()
{
	trace_printk("SIH B B\n");
	return 0;
}

unsigned int _version SEC("version") = 0xFFFFFFFE;

char _license[] SEC("license") = "GPL";