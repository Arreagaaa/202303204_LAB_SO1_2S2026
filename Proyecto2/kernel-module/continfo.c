#include <linux/init.h>
#include <linux/kernel.h>
#include <linux/mm.h>
#include <linux/module.h>
#include <linux/proc_fs.h>
#include <linux/sched/signal.h>
#include <linux/seq_file.h>
#include <linux/sizes.h>
#include <linux/timekeeping.h>

#define PROC_NAME "continfo_pr2_so1_202303204"

MODULE_LICENSE("GPL");
MODULE_AUTHOR("202303204");
MODULE_DESCRIPTION("Métricas de memoria y procesos para Proyecto 2 de SO1");

static struct proc_dir_entry *proc_entry;

static unsigned long process_memory_percent(struct task_struct *task,
                                            unsigned long total_pages)
{
    unsigned long resident_pages;

    if (!task->mm || total_pages == 0)
        return 0;

    resident_pages = get_mm_rss(task->mm);
    return (resident_pages * 10000UL) / total_pages;
}

static unsigned long process_cpu_percent(struct task_struct *task)
{
    u64 runtime_ns = task->se.sum_exec_runtime;
    u64 elapsed_ns = ktime_get_ns() - task->start_time;

    if (elapsed_ns == 0)
        return 0;

    return (runtime_ns * 10000ULL) / elapsed_ns;
}

static int continfo_show(struct seq_file *seq, void *unused)
{
    struct sysinfo info;
    struct task_struct *task;
    unsigned long total_pages;
    unsigned long total_kb;
    unsigned long free_kb;
    unsigned long used_kb;

    si_meminfo(&info);
    total_pages = info.totalram;
    total_kb = (total_pages * (PAGE_SIZE / SZ_1K));
    free_kb = (info.freeram * (PAGE_SIZE / SZ_1K));
    used_kb = total_kb > free_kb ? total_kb - free_kb : 0;

    seq_printf(seq, "MEM_TOTAL_KB:%lu\n", total_kb);
    seq_printf(seq, "MEM_FREE_KB:%lu\n", free_kb);
    seq_printf(seq, "MEM_USED_KB:%lu\n", used_kb);
    seq_puts(seq, "PID,NAME,CMD_OR_CONTAINER_ID,VSZ_KB,RSS_KB,MEM_PCT,CPU_PCT,CLASS\n");

    rcu_read_lock();
    for_each_process(task) {
        unsigned long vsz_kb = 0;
        unsigned long rss_kb = 0;
        unsigned long mem_pct = 0;
        unsigned long cpu_pct = process_cpu_percent(task);

        if (task->mm) {
            vsz_kb = (task->mm->total_vm * (PAGE_SIZE / SZ_1K));
            rss_kb = (get_mm_rss(task->mm) * (PAGE_SIZE / SZ_1K));
            mem_pct = process_memory_percent(task, total_pages);
        }

        seq_printf(seq, "%d,%s,%s,%lu,%lu,%lu.%02lu,%lu.%02lu,unknown\n",
                   task_pid_nr(task), task->comm, task->comm, vsz_kb, rss_kb,
                   mem_pct / 100, mem_pct % 100, cpu_pct / 100, cpu_pct % 100);
    }
    rcu_read_unlock();
    return 0;
}

static int continfo_open(struct inode *inode, struct file *file)
{
    return single_open(file, continfo_show, NULL);
}

static const struct proc_ops continfo_ops = {
    .proc_open = continfo_open,
    .proc_read = seq_read,
    .proc_lseek = seq_lseek,
    .proc_release = single_release,
};

static int __init continfo_init(void)
{
    proc_entry = proc_create(PROC_NAME, 0444, NULL, &continfo_ops);
    if (!proc_entry)
        return -ENOMEM;

    pr_info("continfo: módulo cargado, archivo /proc/%s\n", PROC_NAME);
    return 0;
}

static void __exit continfo_exit(void)
{
    proc_remove(proc_entry);
    pr_info("continfo: módulo descargado\n");
}

module_init(continfo_init);
module_exit(continfo_exit);
