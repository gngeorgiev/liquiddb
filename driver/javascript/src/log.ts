import * as colors from 'colors/safe';

export enum LogLevel {
    off,
    fatal,
    error,
    warning,
    information,
    debug,
    verbose
}

const logColors = {
    [LogLevel.fatal]: colors.red,
    [LogLevel.error]: colors.red,
    [LogLevel.warning]: colors.yellow,
    [LogLevel.information]: colors.blue,
    [LogLevel.debug]: colors.cyan,
    [LogLevel.verbose]: colors.white
};

let logLevel: LogLevel = LogLevel.off;

export function logger(component: string) {
    const log = (data: any, level: LogLevel) => {
        const levelNumber = LogLevel[logLevel] as any;
        if (level > levelNumber) {
            return;
        }

        const msg = typeof data === 'string' ? data : JSON.stringify(data);

        const color = logColors[level];
        const levelString = LogLevel[level];
        console.log(`${color(levelString)} ${component} | ${msg}`);
    };

    return {
        fatal(msg: any) {
            log(msg, LogLevel.fatal);
        },
        error(msg: any) {
            log(msg, LogLevel.error);
        },
        warning(msg: any) {
            log(msg, LogLevel.warning);
        },
        info(msg: any) {
            log(msg, LogLevel.information);
        },
        debug(msg: any) {
            log(msg, LogLevel.debug);
        },
        verbose(msg: any) {
            log(msg, LogLevel.verbose);
        }
    };
}

export function configure({ level }: { level: LogLevel }) {
    logLevel = level;
}
