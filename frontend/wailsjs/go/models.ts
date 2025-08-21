export namespace schema {
	
	export class GroupVersionResource {
	    Group: string;
	    Version: string;
	    Resource: string;
	
	    static createFrom(source: any = {}) {
	        return new GroupVersionResource(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Group = source["Group"];
	        this.Version = source["Version"];
	        this.Resource = source["Resource"];
	    }
	}

}

