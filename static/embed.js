// vim: set ts=2 sw=2 et :
'use strict';
window.addEventListener('DOMContentLoaded', function() {
	var commentElt = document.getElementById('ss-comments');
	if(commentElt === null) {
		console.error('"#ss-comments" element not found; can\'t embed comments.');
		return;
	}

	var apiInfo = function() {
		var scriptElt = document.getElementById('ss-comments-script');
		var baseUrl = scriptElt.attributes.src.value.replace('/static/embed.min.js', '');
		var parts = baseUrl.split('/');
		var key = parts[parts.length - 1]; // api key; this gives us general
																				 // access to the grain.
		var proto = parts[0];
		var host = parts[2]; // Two slashes in the http://, then our host.
		var domainparts = host.split('.');
		var subdomain = domainparts[0];
		var maindomain = new Array(domainparts.length - 1);
    for(var i = 0; i < maindomain.length; i++) {
      maindomain[i] = domainparts[i+1];
    }
		return {
			key: key,
			subdomain: subdomain,
			maindomain: maindomain.join('.'),
			proto: proto,
			baseUrl: baseUrl,
		}
	}();

	var articleId = window.location.host + window.location.pathname;
	var commentUrl = apiInfo.baseUrl + '/comments?articleId=' +
		encodeURIComponent(articleId);

	var req = new XMLHttpRequest();
	req.onreadystatechange = function() {
		if(req.readyState === XMLHttpRequest.DONE) {
			if(req.status !== 200) {
				console.error('Error fetching comments; server returned non-200 status');
				return;
			}
      var substitutions = {
				'<${KEY}/>': apiInfo.key,
				'<${MAIN_DOMAIN}/>': apiInfo.maindomain,
				'<${SUB_DOMAIN}/>': apiInfo.subdomain,
				'<${PROTO}/>': apiInfo.proto,
				'<${BASE_URL}/>': apiInfo.baseUrl,
				'<${REDIRECT_RAW}/>': window.location.href,
				'<${REDIRECT_URI}/>': encodeURIComponent(window.location.href),
      }
      var text = req.responseText
      for (var key in substitutions) {
        text = text.split(key).join(substitutions[key]);
      }
      commentElt.innerHTML = text;
		}
	};
	req.open('GET', commentUrl, true);
	req.send();
});
