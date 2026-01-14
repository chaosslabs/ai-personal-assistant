export namespace models {
	
	export class Activity {
	    id: string;
	    user_id: string;
	    type: string;
	    title: string;
	    // Go type: time
	    start_time: any;
	    // Go type: time
	    end_time?: any;
	    status: string;
	    tags: string[];
	    metadata: Record<string, any>;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	    // Go type: time
	    deleted_at?: any;
	
	    static createFrom(source: any = {}) {
	        return new Activity(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.user_id = source["user_id"];
	        this.type = source["type"];
	        this.title = source["title"];
	        this.start_time = this.convertValues(source["start_time"], null);
	        this.end_time = this.convertValues(source["end_time"], null);
	        this.status = source["status"];
	        this.tags = source["tags"];
	        this.metadata = source["metadata"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	        this.deleted_at = this.convertValues(source["deleted_at"], null);
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
	export class AudioDeviceInfo {
	    name: string;
	    device_id?: string;
	    sample_rate: number;
	    channels: number;
	    bit_depth?: number;
	    device_type?: string;
	
	    static createFrom(source: any = {}) {
	        return new AudioDeviceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.device_id = source["device_id"];
	        this.sample_rate = source["sample_rate"];
	        this.channels = source["channels"];
	        this.bit_depth = source["bit_depth"];
	        this.device_type = source["device_type"];
	    }
	}
	export class RecordingConfig {
	    format: string;
	    quality: string;
	    sample_rate: number;
	    bitrate?: number;
	    auto_gain_control: boolean;
	    noise_reduction: boolean;
	    chunk_size?: number;
	    recording_mode: string;
	
	    static createFrom(source: any = {}) {
	        return new RecordingConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.format = source["format"];
	        this.quality = source["quality"];
	        this.sample_rate = source["sample_rate"];
	        this.bitrate = source["bitrate"];
	        this.auto_gain_control = source["auto_gain_control"];
	        this.noise_reduction = source["noise_reduction"];
	        this.chunk_size = source["chunk_size"];
	        this.recording_mode = source["recording_mode"];
	    }
	}
	export class AudioRecording {
	    id: string;
	    user_id: string;
	    activity_id: string;
	    file_path: string;
	    device_info: AudioDeviceInfo;
	    status: string;
	    duration?: number;
	    file_size?: number;
	    config: RecordingConfig;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new AudioRecording(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.user_id = source["user_id"];
	        this.activity_id = source["activity_id"];
	        this.file_path = source["file_path"];
	        this.device_info = this.convertValues(source["device_info"], AudioDeviceInfo);
	        this.status = source["status"];
	        this.duration = source["duration"];
	        this.file_size = source["file_size"];
	        this.config = this.convertValues(source["config"], RecordingConfig);
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	
	export class TranscriptChunk {
	    id: string;
	    user_id: string;
	    activity_id: string;
	    audio_recording_id: string;
	    text: string;
	    start_time: number;
	    end_time: number;
	    speaker?: string;
	    confidence?: number;
	    language?: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new TranscriptChunk(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.user_id = source["user_id"];
	        this.activity_id = source["activity_id"];
	        this.audio_recording_id = source["audio_recording_id"];
	        this.text = source["text"];
	        this.start_time = source["start_time"];
	        this.end_time = source["end_time"];
	        this.speaker = source["speaker"];
	        this.confidence = source["confidence"];
	        this.language = source["language"];
	        this.created_at = this.convertValues(source["created_at"], null);
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
	export class TranscriptionStatus {
	    stage: string;
	    progress: number;
	    processed_chunks: number;
	    total_chunks: number;
	    current_file: string;
	    estimated_time: number;
	    // Go type: time
	    started_at: any;
	    last_error?: string;
	
	    static createFrom(source: any = {}) {
	        return new TranscriptionStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stage = source["stage"];
	        this.progress = source["progress"];
	        this.processed_chunks = source["processed_chunks"];
	        this.total_chunks = source["total_chunks"];
	        this.current_file = source["current_file"];
	        this.estimated_time = source["estimated_time"];
	        this.started_at = this.convertValues(source["started_at"], null);
	        this.last_error = source["last_error"];
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
	export class WhisperModel {
	    id: string;
	    name: string;
	    size: number;
	    is_downloaded: boolean;
	    is_active: boolean;
	    languages: string[];
	    accuracy: string;
	    speed: string;
	
	    static createFrom(source: any = {}) {
	        return new WhisperModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.size = source["size"];
	        this.is_downloaded = source["is_downloaded"];
	        this.is_active = source["is_active"];
	        this.languages = source["languages"];
	        this.accuracy = source["accuracy"];
	        this.speed = source["speed"];
	    }
	}

}

export namespace views {
	
	export class RecordingSession {
	    activity?: models.Activity;
	    audio_recording?: models.AudioRecording;
	    file_path: string;
	
	    static createFrom(source: any = {}) {
	        return new RecordingSession(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.activity = this.convertValues(source["activity"], models.Activity);
	        this.audio_recording = this.convertValues(source["audio_recording"], models.AudioRecording);
	        this.file_path = source["file_path"];
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

