export function replaceLinks(body: string): string {
    return body.replace(/\[\[(.*?)\]\]/g, (match, p1) => {
        return `<a href="#/doc-title/${p1}">${p1}</a>`;
    });
}