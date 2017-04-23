'use strict';
window.addEventListener('DOMContentLoaded', function() {
	var commentElt = document.getElementById('ss-comments');
	if(commentElt === null) {
		console.error('"#ss-comments" element not found; can\'t embed comments.');
		return;
	}

	var scriptElt = document.getElementById('ss-comments-script');
	var articleId = window.location.host + window.location.pathname;
	var baseUrl = scriptElt.attributes.src.value.replace('static/embed.min.js', '');
	var commentUrl = baseUrl + 'comments?articleId=' +
		encodeURIComponent(articleId);

	var req = new XMLHttpRequest();
	req.onreadystatechange = function() {
		if(req.readyState === XMLHttpRequest.DONE) {
			if(req.status !== 200) {
				console.error('Error fetching comments; server returned non-200 status');
				return;
			}
			commentElt.innerHTML = req.responseText;
			var formElt = commentElt.getElementsByTagName('form')[0];
			formElt.attributes.action.value = baseUrl + 'new-comment';
			var redirectElt = document.getElementById('ss-comments-redirect');
			redirectElt.attributes.value.value = window.location.href;
		}
	};
	req.open('GET', commentUrl, true);
	req.send();
});
