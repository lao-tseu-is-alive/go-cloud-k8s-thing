/**
 * check if the given variable is null or undefined
 * @param variable
 * @returns true if given variable is null or undefined, false in other cases
 */
export const isNullOrUndefined = (variable: any): boolean => typeof variable === 'undefined' || variable === null;
