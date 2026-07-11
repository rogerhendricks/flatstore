export namespace flathub {
	
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

