export declare enum LogLevel {
    off = 0,
    fatal = 1,
    error = 2,
    warning = 3,
    information = 4,
    debug = 5,
    verbose = 6,
}
export declare function logger(component: string): {
    fatal(msg: any): void;
    error(msg: any): void;
    warning(msg: any): void;
    information(msg: any): void;
    debug(msg: any): void;
    verbose(msg: any): void;
};
export declare function configure({level}: {
    level: LogLevel;
}): void;
