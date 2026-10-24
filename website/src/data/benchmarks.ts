// Results behind the benchmark charts and tables, updated together. The suite reruns with the commands on the
// Benchmarks page; the comparison comes from a harness that runs the old and new versions side by side.

export const environment = 'Apple M4 Max, Go 1.27.0, median of 5 runs';

/** The two versions the charts compare. */
export const versions = { before: 'v1', after: 'v2' } as const;

export type Group = 'publish' | 'parallel' | 'subscribe';

export interface Result {
	ns: number;
	allocs: number;
}

export interface Comparison {
	scenario: string;
	before: Result;
	after: Result;
}

/**
 * v1 against v2 on the same harness, one call per op. Queues hold 1024 messages and handlers count what they get,
 * so what is measured is the bus's own cost: queueing, waking the handlers, and waiting when they fall behind.
 */
export const comparison: Record<Group, Comparison[]> = {
	publish: [
		{ scenario: '1 subscriber', before: { ns: 190.9, allocs: 2 }, after: { ns: 48.6, allocs: 0 } },
		{ scenario: '10 subscribers', before: { ns: 895, allocs: 2 }, after: { ns: 449.3, allocs: 0 } },
		{ scenario: '100 subscribers', before: { ns: 13020, allocs: 2 }, after: { ns: 9810, allocs: 0 } },
		{ scenario: '100 topics, 1 subscriber each', before: { ns: 146.8, allocs: 2 }, after: { ns: 112.2, allocs: 0 } },
	],
	parallel: [
		{ scenario: '1 subscriber', before: { ns: 347.5, allocs: 2 }, after: { ns: 109.4, allocs: 0 } },
		{ scenario: '100 topics, 1 subscriber each', before: { ns: 159.2, allocs: 2 }, after: { ns: 27.7, allocs: 0 } },
	],
	subscribe: [{ scenario: 'Subscribe, then unsubscribe', before: { ns: 2450, allocs: 5 }, after: { ns: 1120, allocs: 6 } }],
};

export interface SuiteResult {
	benchmark: string;
	what: string;
	ns: number;
	bytes: number;
	allocs: number;
}

/** v2's own benchmark suite, `go test -bench .`. */
export const suite: SuiteResult[] = [
	{ benchmark: 'Publish/1', what: 'One subscriber', ns: 42.8, bytes: 0, allocs: 0 },
	{ benchmark: 'Publish/10', what: 'Ten subscribers to one topic', ns: 423.7, bytes: 0, allocs: 0 },
	{ benchmark: 'Publish/100', what: 'A hundred subscribers to one topic', ns: 8670, bytes: 0, allocs: 0 },
	{ benchmark: 'PublishTopics', what: 'A hundred topics in turn, one subscriber each', ns: 103.9, bytes: 0, allocs: 0 },
	{ benchmark: 'PublishUnbuffered', what: 'New(0): each message is handed straight to the handler', ns: 155.5, bytes: 0, allocs: 0 },
	{ benchmark: 'PublishParallel', what: 'Publish/1 from 16 goroutines at once', ns: 102, bytes: 0, allocs: 0 },
	{ benchmark: 'PublishParallelTopics', what: 'PublishTopics from 16 goroutines at once', ns: 27.6, bytes: 0, allocs: 0 },
	{ benchmark: 'Subscribe', what: 'Subscribe, then unsubscribe', ns: 1090, bytes: 9736, allocs: 6 },
];
