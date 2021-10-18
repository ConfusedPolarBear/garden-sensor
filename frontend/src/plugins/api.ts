export default function api(
  url: string,
  options: RequestInit = {}
): Promise<Response> {
  const addr = window.localStorage.getItem("server");

  if (!addr) {
    throw new Error("Address is null");
  }

  // Construct the URL
  let where = addr;
  if (where.endsWith("/")) {
    where = where.substr(0, where.length - 1);
  }
  where += url;

  // Send it
  console.debug(`[api] fetching ${where}`);
  return fetch(where, options);
}
