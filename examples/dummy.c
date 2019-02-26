#include "elf/include/bpf.h"
#include "elf/include/bpf_map.h"
#include "elf/include/bpf_helpers.h"
#include "elf/include/trace_printk.h"

struct bpf_map_def SEC("maps/sys_enter_write") dummy = {
	.type = BPF_MAP_TYPE_HASH,
	.key_size = sizeof(int),
	.value_size = sizeof(unsigned int),
	// cat /proc/sys/kernel/pid_max 
	// normative practice on most 64 bit systems to set this value to the same value as found on 32 bit systems
	.max_entries = 32768, 
	// .pinning = PIN_GLOBAL_NS,
	// .namespace = 'ns',
};

SEC("tracepoint/syscalls/sys_enter_write")
int tracepoint__syscalls__sys_enter_write()
{
	int id = bpf_get_current_pid_tgid();

	// struct cnt * counter = bpf_map_lookup_elem(&dummy_hash, &id);
	// if (counter) {
	// 	bpf_spin_lock(&counter->lock);
	// 	counter->val++;
	// 	bpf_spin_unlock(&counter->lock);
	// }
	
	int one = 1;
	int *el = bpf_map_lookup_elem(&dummy, &id);
	if (el) {
		(*el)++;
	} else {
		el = &one;
	}

	// todos > spin lock to mutex a struct containing the value (serialize access to it)
	bpf_map_update_elem(&dummy, &id, el, BPF_ANY);

	return 0;
}

char _license[] SEC("license") = "GPL";

unsigned int _version SEC("version") = 0xFFFFFFFE; // this tells to the ELF loader to set the current running kernel version
