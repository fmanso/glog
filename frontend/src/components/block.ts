export class Block {
    id: string
    content: string
    indent: number

    constructor(id: string, content: string, indent: number = 0) {
        this.id = id
        this.content = content
        this.indent = indent
    }
}

export class Document {
    id: string
    title: string
    blocks: Block[]

    constructor(id: string, title: string) {
        this.id = id
        this.title = title
        this.blocks = []
    }
}