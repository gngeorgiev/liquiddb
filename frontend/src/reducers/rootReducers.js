export const db = (state = {}, action) => {
    if (action.type === 'INITIALIZE_DB') {
        return action.db;
    }

    return state;
};

export const dbData = (state = {}, action) => {
    if (action.type === 'DB_DATA') {
        return action.data || {};
    }

    return state;
};

export const dbStats = (state = {}, action) => {
    if (action.type === 'INITIALIZE_DB_STATS') {
        return action.stats;
    }

    return state;
};

export const dbStatsData = (state = {}, action) => {
    if (action.type === 'DB_STATS_DATA') {
        return action.data;
    }

    return state;
};
