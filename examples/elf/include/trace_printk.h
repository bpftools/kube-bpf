#ifndef __UTILS__
#define __UTILS__

// BPF_FUNC_trace_printk is a member of an enum, its value is 6
static int (*bpf_trace_printk)(const char *fmt, int fmt_size, ...) = (void *)BPF_FUNC_trace_printk;

#define trace_printk(fmt, ...)                                                 \
  do {                                                                         \
    char _fmt[] = fmt;                                                         \
    bpf_trace_printk(_fmt, sizeof(_fmt), ##__VA_ARGS__);                       \
  } while (0)

#endif