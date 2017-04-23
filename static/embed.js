'use strict';
window.addEventListener('DOMContentLoaded', function() {
	var articleId = window.location.host + window.location.pathname;
	var commentElt = document.getElementById('comments');
	var commentUrl = 'http://localhost:8899/comments?articleId=' +
		encodeURIComponent(articleId);

	if(commentElt === null) {
		console.error('"#comment" element not found; can\'t embed comments.');
		return;
	}

	var req = new XMLHttpRequest();
	req.onreadystatechange = function() {
		if(req.readyState === XMLHttpRequest.DONE) {
			if(req.status !== 200) {
				console.error('Error fetching comments; server returned non-200 status');
				return;
			}
			commentElt.innerHTML = req.responseText;
		}
	};
	req.open('GET', commentUrl, true);
	req.send();
});
