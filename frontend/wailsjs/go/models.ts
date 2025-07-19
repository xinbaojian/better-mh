export namespace adb {
	
	export class Adb {
	    Connected: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Adb(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Connected = source["Connected"];
	    }
	}

}

export namespace gocv {
	
	export class Mat {
	
	
	    static createFrom(source: any = {}) {
	        return new Mat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

export namespace message {
	
	export class Message {
	
	
	    static createFrom(source: any = {}) {
	        return new Message(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

