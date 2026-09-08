export class Preferences {
  constructor(props: { theme: string; themeId: string; zoomLevel: number; showHidden: boolean; launchAtLogin: boolean }) {
    Object.assign(this, props);
  }
  theme: string = "";
  themeId: string = "";
  zoomLevel: number = 0;
  showHidden: boolean = false;
  launchAtLogin: boolean = false;
}

export class SearchResult {
  constructor(props: { path: string; isDir: boolean; size: number }) {
    Object.assign(this, props);
  }
  path: string = "";
  isDir: boolean = false;
  size: number = 0;
}

export class VersionInfo {
  constructor(props: { version: string; commit: string; buildTime: string }) {
    Object.assign(this, props);
  }
  version: string = "";
  commit: string = "";
  buildTime: string = "";
}

export class SearchOptions {
  constructor(props: { root?: string; fileTypes?: string[]; skipDirs?: string[]; includeHidden?: boolean }) {
    Object.assign(this, props);
  }
  root?: string;
  fileTypes?: string[];
  skipDirs?: string[];
  includeHidden?: boolean;
}

export class Release {
  constructor(props: { version: string; notes: string; url: string }) {
    Object.assign(this, props);
  }
  version: string = "";
  notes: string = "";
  url: string = "";
}
