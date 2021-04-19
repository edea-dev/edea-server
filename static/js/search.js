const search = instantsearch({
    indexName: "edea",
    searchClient: instantMeiliSearch(
        "http://localhost:7700"
    )
    });

    search.addWidgets([
      instantsearch.widgets.searchBox({
          container: "#searchbox"
      }),
      instantsearch.widgets.configure({ hitsPerPage: 8 }),
      instantsearch.widgets.hits({
          container: "#hits",
          templates: {
          item: `
              <div class="card">
                <div class="card-body">
                <h5 class="hit-name">
                  {{#helpers.highlight}}{ "attribute": "Name" }{{/helpers.highlight}}
                </h5>
                <h6 class="card-subtitle mb-2 text-muted">by \{{Author}}</h6>
                <p class="card-text hit-description">
                  {{#helpers.highlight}}{ "attribute": "Description" }{{/helpers.highlight}}
                </p>
                <a href="/\{{Type}}/\{{ID}}" class="card-link">View</a>
                <a href="#" class="card-link">Card link</a>
                </div>
              </div>
          `
          }
      })
    ]);
search.start();
