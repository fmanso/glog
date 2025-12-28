export namespace main {
	
	export class BlockDto {
	    id: string;
	    content: string;
	    indent: number;
	
	    static createFrom(source: any = {}) {
	        return new BlockDto(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.content = source["content"];
	        this.indent = source["indent"];
	    }
	}
	export class DocumentDto {
	    id: string;
	    title: string;
	    blocks: BlockDto[];
	    date: string;
	
	    static createFrom(source: any = {}) {
	        return new DocumentDto(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.blocks = this.convertValues(source["blocks"], BlockDto);
	        this.date = source["date"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DocumentSummaryDto {
	    id: string;
	    title: string;
	    date: string;
	
	    static createFrom(source: any = {}) {
	        return new DocumentSummaryDto(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.date = source["date"];
	    }
	}

}

