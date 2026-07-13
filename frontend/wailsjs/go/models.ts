export namespace flathub {
	
	export class AppDetails {
	    flatpakAppId: string;
	    name: string;
	    summary: string;
	    description: string;
	    homepageUrl: string;
	    bugtrackerUrl: string;
	    iconUrl: string;
	    version: string;
	    developer: string;
	    screenshots: string[];
	    releaseDate: string;
	
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
	        this.iconUrl = source["iconUrl"];
	        this.version = source["version"];
	        this.developer = source["developer"];
	        this.screenshots = source["screenshots"];
	        this.releaseDate = source["releaseDate"];
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

