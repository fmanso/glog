export function replaceLinks(body: string): string {
    return body.replace(/\[\[(.*?)\]\]/g, (match, p1) => {
        // URL-encode the title to handle slashes and other special characters
        const encodedTitle = encodeURIComponent(p1);
        return `<a href="#/doc-title/${encodedTitle}">${p1}</a>`;
    });
}