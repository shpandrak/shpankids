
export function descriptionToKebabCase(description: string): string {
    return description
        .toLowerCase() // Convert the string to lowercase
        .replace(/\s+/g, '-') // Replace spaces with hyphens
        .replace(/[^a-z0-9-]/g, '') // Remove any characters that are not letters, numbers, or hyphens
        .replace(/--+/g, '-') // Replace multiple hyphens with a single hyphen
        .trim(); // Trim any leading or trailing spaces
}


// eslint-disable-next-line
export function showError(reason: any, prefix: string = "") {
    console.error(reason);
    if (!reason?.response || !(reason.response instanceof Response)) {
        alert(prefix + reason);
    }
    (reason.response as Response)
        .text()
        .then((data) => {
            alert(prefix + (reason.response as Response).statusText + ': ' + data);
        })
}
