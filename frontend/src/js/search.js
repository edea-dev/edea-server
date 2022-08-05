
// MeiliSearch
const searchClient = new MeiliSearch({
	host: 'http://192.168.0.2:7700',
	apiKey: '',
})

const fp_index = searchClient.index('edea')

async function searchbox_handler() {
	var element = document.querySelector("input[name='searchbox']")
	const search = await fp_index.search(element.value)

	if (search.hits.length > 0) {
		var hits = document.querySelector("table[id='hits'] tbody")

		console.log(search.hits);

		var body = document.createElement('tbody');

		search.hits.forEach(e => {
			let row = body.insertRow(-1)
			let type = row.insertCell(0)
			let name = row.insertCell(1)
			let author = row.insertCell(2)
			let desc = row.insertCell(3)

			var link = document.createElement('a');
			link.innerHTML = e.Name

			if (e.Type == "module") {
				link.href = "/module/" + e.ID
			} else if (e.Type == "bench") {
				link.href = "/bench/" + e.ID
			}

			name.appendChild(link)
			type.innerHTML = e.Type
			author.innerHTML = e.Author
			desc.innerHTML = e.Description
		});

		// update the content with out search results
		hits.parentNode.replaceChild(body, hits)
	}
}

// if we have a query param, fill it out
var searchParams = new URLSearchParams(window.location.search);
if (searchParams.has('q')) {
	var element = document.querySelector("input[name='searchbox']");
	element.value = searchParams.get('q');
}

searchbox_handler();

async function categories() {
	const categories = await fetch(`/api/search_fields`).then((response) => response.json())
	var filter_table = document.querySelector("table[id='filters'] tbody")
	var body = document.createElement('tbody');

	const cats = Object.keys(categories)

	if (cats.length > 0) {
		let row = body.insertRow(-1)
		let i = 0;
		cats.forEach(cat => {
			let cell = row.insertCell(i)
			i++;

			let label = document.createElement('label')
			label.innerText = cat

			cell.appendChild(label)

			let select = document.createElement('select')
			select.name = cat
			select.setAttribute("multiple", "")

			let optgroup = document.createElement('optgroup')

			// add an option for each value
			categories[cat].forEach(value => {
				let opt = document.createElement('option')
				opt.value = value
				opt.innerText = value
				optgroup.appendChild(opt)
			})

			// add everything together
			select.appendChild(optgroup)
			cell.appendChild(select)

			// TODO: add some more controls to it
		})
	}

	filter_table.parentNode.replaceChild(body, filter_table)
}

categories();
