# frozen_string_literal: true

class NavbarComponent < ApplicationComponent
  def self.template
    <<~ERB
      <nav class="navbar navbar-expand-lg navbar-light bg-white fixed-top shadow-sm">
        <button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbar-toggler">
          <span class="navbar-toggler-icon"></span>
        </button>
        <div class="collapse navbar-collapse" id="navbar-toggler">
          <%= link_to root_path, class: "navbar-brand" do %>
            <%= image_tag asset_bundle_path("images/logo-color-mizuho.png"), size: "25x30", alt: "Annict" %>
          <% end %>
          <ul class="navbar-nav mr-auto mt-2 mt-lg-0">
            <li class="nav-item dropdown active">
              <a class="nav-link dropdown-toggle" href="" data-toggle="dropdown">
                <%= t("verb.explore") %>
              </a>
              <div class="dropdown-menu">
                <%= link_to t("noun.current_season"), "", class: "dropdown-item" %>
                <%= link_to t("noun.next_season"), "", class: "dropdown-item" %>
                <%= link_to t("noun.prev_season"), "", class: "dropdown-item" %>
                <%= link_to t("verb.search"), "", class: "dropdown-item" %>
              </div>
            </li>
            <li class="nav-item">
              <a class="nav-link" href="#">Link</a>
            </li>
            <li class="nav-item">
              <a class="nav-link disabled" href="#" tabindex="-1" aria-disabled="true">Disabled</a>
            </li>
          </ul>
          <form class="form-inline my-2 my-lg-0">
            <input class="form-control mr-sm-2" type="search" placeholder="Search" aria-label="Search">
            <button class="btn btn-outline-success my-2 my-sm-0" type="submit">Search</button>
          </form>
        </div>
      </nav>
    ERB
  end
end
