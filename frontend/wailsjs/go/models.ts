export namespace installer {
	
	export class Result {
	    rootPath: string;
	    target: string;
	    platform: string;
	    finishedAt: string;
	    launcher: string;
	    verified: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Result(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rootPath = source["rootPath"];
	        this.target = source["target"];
	        this.platform = source["platform"];
	        this.finishedAt = source["finishedAt"];
	        this.launcher = source["launcher"];
	        this.verified = source["verified"];
	    }
	}

}

