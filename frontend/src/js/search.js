// simple fulltext search

async function searchbox_handler(event) {
	var element
	if (event === undefined) {
		element = document.getElementById("searchbox")
	} else {
		element = event.target
	}
	const formData = new FormData();
	formData.append('q', element.value)

	const search = await fetch(
		"/search", {
		method: 'POST',
		body: formData,
		headers: {
			'Accept': 'application/json',
		},
	}).then((r) => {
		if (r.status == 200) {
			return r.json()
		} else {
			return {hits: []}
		}
	})

	if (search.hits.length > 0) {
		var hits = document.querySelector("table[id='hits-table'] tbody")

		var body = document.createElement('tbody');

		search.hits.forEach(e => {
			let row = body.insertRow(-1)
			let type = row.insertCell(0)
			let name = row.insertCell(1)
			let author = row.insertCell(2)
			let desc = row.insertCell(3)

			var link = document.createElement('a');
			link.innerHTML = e._formatted.name

			if (e.type == "module") {
				link.href = "/module/" + e.id
			} else if (e.type == "bench") {
				link.href = "/bench/" + e.id
			}

			name.appendChild(link)
			type.innerHTML = e._formatted.type
			author.innerHTML = e._formatted.author
			desc.innerHTML = e._formatted.description
		});

		// update the content with out search results
		hits.parentNode.replaceChild(body, hits)
	}
}

// if we have a query param, fill it out
window.addEventListener('DOMContentLoaded', (event) => {
	var searchParams = new URLSearchParams(window.location.search);
	if (searchParams.has('q')) {
		var element = document.querySelector("input[id='searchbox']");
		element.value = searchParams.get('q');
		searchbox_handler();
	}

	const input = document.getElementById('searchbox');
	input.addEventListener('input', searchbox_handler);
});
