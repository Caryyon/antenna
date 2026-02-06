export namespace main {
	
	export class Session {
	    sessionId: string;
	    name: string;
	    kind: string;
	    model: string;
	    messageCount: number;
	    totalCost: number;
	    todayCost: number;
	    updatedAt: number;
	    isActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Session(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.model = source["model"];
	        this.messageCount = source["messageCount"];
	        this.totalCost = source["totalCost"];
	        this.todayCost = source["todayCost"];
	        this.updatedAt = source["updatedAt"];
	        this.isActive = source["isActive"];
	    }
	}
	export class DashboardData {
	    sessions: Session[];
	    totalCount: number;
	    totalCost: number;
	    todayCost: number;
	
	    static createFrom(source: any = {}) {
	        return new DashboardData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessions = this.convertValues(source["sessions"], Session);
	        this.totalCount = source["totalCount"];
	        this.totalCost = source["totalCost"];
	        this.todayCost = source["todayCost"];
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

}

