
// MeiliSearch
const searchClient = new MeiliSearch({
	host: 'http://pinus:7700',
	apiKey: 'meiliedea',
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

function select_range_handler(event) {
	event.preventDefault()

	let btn = event.srcElement
	let select_element = btn.parentElement.parentElement.children[1]

	if (btn.attributes["aria-label"].value.includes("less than")) {
		for (var i = 0; i < select_element.options.length; i++) {
			if (! select_element.options[i].selected) {
				select_element.options[i].selected = true
			} else {
				break
			}
		}
	} else if (btn.attributes["aria-label"].value.includes("larger than")) {
		for (var i = select_element.options.length - 1; i >= 0; i--) {
			if (! select_element.options[i].selected) {
				select_element.options[i].selected = true
			} else {
				break
			}
		}
	} else if (btn.attributes["aria-label"].value.includes("clear")) {
		for (var i = 0; i < select_element.options.length; i++) {
			if (select_element.options[i].selected) {
				select_element.options[i].selected = false
			}
		}
		btn.disabled = true
		btn.blur()
	}
}


function _create_button(button_label, aria_label, color, disabled=false) {
	let btn = document.createElement("button")
	btn.classList.add("btn")
	btn.setAttribute("aria-label", aria_label)
	btn.innerHTML = button_label
	if (! disabled) {
		btn.addEventListener('click', enable_update_filters_btn)
		btn.addEventListener('click', select_range_handler)
		btn.classList.add("btn-outline-" + color)
	} else {
		btn.classList.add("btn-outline-light")
		btn.setAttribute("disabled", true)
	}
	return btn
}

const filterfield_prefix = "filterf_"

async function categories() {
	const categories = await fetch(`/api/search_fields`).then((response) => response.json())
	var filters_container = document.getElementById("filters-row")

	const cats = Object.keys(categories)

	if (cats.length > 0) {

		let i = 0;
		cats.forEach(cat => {
			let outer_div = document.createElement("div")
			outer_div.classList.add("filterbox", "col-3")  // change container width here when necessary

			let form_control_element_name = filterfield_prefix + cat

			let label = document.createElement('label')
			label.innerText = cat  // TODO add human readable category name here
			label.setAttribute("for", form_control_element_name)
			label.classList.add("form-label")

			outer_div.appendChild(label)

			let select = document.createElement('select')
			select.name = form_control_element_name
			select.setAttribute("multiple", "")
			select.setAttribute("aria-label", "filter options for " + cat)
			select.classList.add("form-select", "form-select-sm", "mb-1")
			select.addEventListener('change', enable_update_filters_btn)

			// add an option for each value
			var num_cats = 0
			var num_selected = 0
			categories[cat].forEach(value => {
				let opt = document.createElement('option')
				opt.value = value
				opt.innerText = value
				select.appendChild(opt)
				num_cats++
				if (opt.selected) {
					num_selected++
				}
			})

			// add everything together
			outer_div.appendChild(select)

			let control_div = document.createElement("div")
			control_div.classList.add("input-group", "input-group-sm", "mb-3")

			control_div.appendChild(_create_button("&nbsp;&#x2264;&nbsp;", "select all values less than or equal to selected", "secondary", disabled=(num_cats < 2)))
			let midbtn = _create_button("&nbsp;&#x21bb;&nbsp;", "clear this filter", "primary", disabled=false)
			if (num_selected == 0) {
				midbtn.disabled = true
			}
			midbtn.classList.add("form-control")
			control_div.appendChild(midbtn)
			control_div.appendChild(_create_button("&nbsp;&#x2265;&nbsp;", "select all values larger than or equal to selected", "secondary", disabled=(num_cats < 2)))

			outer_div.appendChild(control_div)

			filters_container.appendChild(outer_div)
			i++
		})
	}

}

function disable_update_filters_btn() {
	update_filters_already_enabled = false
	let filter_button = document.getElementById("filter_apply_btn")
	filter_button.disabled = true
	filter_button.classList.add("btn-outline-light")
	filter_button.classList.remove("btn-primary")
	filter_button.removeEventListener('click', do_search)
}

let update_filters_already_enabled = false

function enable_update_filters_btn(event) {
	// enable reset filter button
	if (event.srcElement.nodeName == "SELECT") {
		event.srcElement.parentElement.lastChild.children[1].disabled = false
	}

	if (update_filters_already_enabled) {
		return
	}
	update_filters_already_enabled = true
	let filter_button = document.getElementById("filter_apply_btn")
	filter_button.disabled = false
	filter_button.classList.remove("btn-outline-light")
	filter_button.classList.add("btn-primary")
	filter_button.addEventListener('click', do_search)
}

categories();
let search_results = []

async function do_search() {
	disable_update_filters_btn()
	var filter_row = document.getElementsByClassName("filterbox")

	var filter_ops = []

	for (var i = 0; i < filter_row.length; i++) {
		let e = filter_row[i].getElementsByTagName('select')[0]
		let collection = e.selectedOptions

		if (collection.length == 0) {
			continue
		}

		var op_values = []

		for (var j = 0; j < collection.length; j++) {
			op_values.push(collection[j].value)
		}

		filter_ops.push({ 'field': e.name.replace(filterfield_prefix, ''), 'op': '=', 'values': op_values })
	}

	const results = await fetch(
		'/api/search_module',
		{ method: 'POST', body: JSON.stringify(filter_ops) }
	).then((response) => response.json())

	search_results = results  // put it into a global for easier debugging from dev console
	let results_container = document.getElementById("hits-row")

	for (var i=0; i < results_container.children.length -1; i++) {
		results_container.removeChild(results_container.lastChild)
	}

	for (var i=0; i < results.length; i++) {
		let r = results[i]
		let result_container_div = document.createElement("div")
		result_container_div.classList.add("col-12", "search-result", "mb-3")

		let result_card = document.createElement("div")
		result_card.classList.add("card")
		result_container_div.appendChild(result_card)

		let result_header = document.createElement("div")
		result_header.classList.add("card-header")
		result_header.innerHTML = "&lt; add tags or categories here maybe? &gt;" // TODO
		result_card.appendChild(result_header)

		let result_body = document.createElement("div")
		result_body.classList.add("card-body")
		result_card.appendChild(result_body)

		let result_title = document.createElement("h5")
		result_title.classList.add("card-title")
		result_title.innerHTML = r.Name
		if (r.User.Handle != "") {
			result_title.innerHTML += " <small> by <i>" + r.User.Handle + "</i></small>"
		}
		result_body.appendChild(result_title)

		let result_text = document.createElement("p")
		result_text.classList.add("card-text")
		result_text.innerHTML = r.Description
		result_body.appendChild(result_text)

		let new_link = document.createElement("a")
		new_link.classList.add("btn", "btn-outline-primary", "btn-sm", "mx-2")
		new_link.setAttribute("role", "button")
		new_link.innerHTML = "bottom text"
		result_body.appendChild(new_link)

		let new_link2 = document.createElement("a")
		new_link2.classList.add("btn", "btn-outline-secondary", "btn-sm", "mx-2")
		new_link.setAttribute("role", "button")
		new_link2.innerHTML = "add to bench"
		result_body.appendChild(new_link2)

		let result_footer = document.createElement("p")
		result_footer.classList.add("card-text", "text-muted", "small", "mt-3")
		result_footer.innerHTML = "BOM: " + r.Metadata.count_part + " parts"
		result_footer.innerHTML += " (" + r.Metadata.count_unique + " unique)"
		result_footer.innerHTML += ", PCB area: &lt;tbd&gt; mm&#xb2;"
		result_footer.innerHTML += ", last updated " + r.UpdatedAt  // TODO make it more human readable
		result_body.appendChild(result_footer)
		
		results_container.appendChild(result_container_div)
	}
}
