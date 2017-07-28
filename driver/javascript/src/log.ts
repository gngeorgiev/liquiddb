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
    const log = (level: LogLevel, data: any) => {
        if (level > logLevel) {
            return;
        }

        let messages = data;
        if (!Array.isArray(data)) {
            messages = [data];
        }

        let message = '';
        messages.forEach(m => {
            let msg;
            if (typeof data === 'string') {
                msg = data;
            } else if (data instanceof Error) {
                msg = data.toString();
            } else {
                msg = JSON.stringify(data);
            }

            message += ` ${msg}`;
        });

        const color = logColors[level];
        const levelString = LogLevel[level];
        console.log(`${color(levelString)} ${component} | ${message}`);
    };

    return {
        fatal(msg: any) {
            log(LogLevel.fatal, msg);
        },
        error(msg: any) {
            log(LogLevel.error, msg);
        },
        warn(msg: any) {
            log(LogLevel.warning, msg);
        },
        info(msg: any) {
            log(LogLevel.information, msg);
        },
        debug(msg: any) {
            log(LogLevel.debug, msg);
        },
        verbose(msg: any) {
            log(LogLevel.verbose, msg);
        }
    };
}

export function configure({ level }: { level: LogLevel }) {
    logLevel = level;
}
