export namespace flathub {
	
	export class Release {
	    version: string;
	    date: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new Release(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.date = source["date"];
	        this.description = source["description"];
	    }
	}
	export class AppDetails {
	    flatpakAppId: string;
	    name: string;
	    summary: string;
	    description: string;
	    homepageUrl: string;
	    bugtrackerUrl: string;
	    helpUrl: string;
	    vcsBrowserUrl: string;
	    iconUrl: string;
	    version: string;
	    developer: string;
	    verified: boolean;
	    screenshots: string[];
	    releases: Release[];
	    releaseDate: string;
	    ageRating: string;
	    license: string;
	
	    static createFrom(source: any = {}) {
	        return new AppDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.flatpakAppId = source["flatpakAppId"];
	        this.name = source["name"];
	        this.summary = source["summary"];
	        this.description = source["description"];
	        this.homepageUrl = source["homepageUrl"];
	        this.bugtrackerUrl = source["bugtrackerUrl"];
	        this.helpUrl = source["helpUrl"];
	        this.vcsBrowserUrl = source["vcsBrowserUrl"];
	        this.iconUrl = source["iconUrl"];
	        this.version = source["version"];
	        this.developer = source["developer"];
	        this.verified = source["verified"];
	        this.screenshots = source["screenshots"];
	        this.releases = this.convertValues(source["releases"], Release);
	        this.releaseDate = source["releaseDate"];
	        this.ageRating = source["ageRating"];
	        this.license = source["license"];
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
	export class AppSummary {
	    flatpakAppId: string;
	    name: string;
	    summary: string;
	    iconUrl: string;
	    version: string;
	    developer: string;
	
	    static createFrom(source: any = {}) {
	        return new AppSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.flatpakAppId = source["flatpakAppId"];
	        this.name = source["name"];
	        this.summary = source["summary"];
	        this.iconUrl = source["iconUrl"];
	        this.version = source["version"];
	        this.developer = source["developer"];
	    }
	}
	export class InstalledApp {
	    appId: string;
	    name: string;
	    version: string;
	    updateAvailable: boolean;
	
	    static createFrom(source: any = {}) {
	        return new InstalledApp(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appId = source["appId"];
	        this.name = source["name"];
	        this.version = source["version"];
	        this.updateAvailable = source["updateAvailable"];
	    }
	}

}

