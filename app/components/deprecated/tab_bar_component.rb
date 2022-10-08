# frozen_string_literal: true

class Deprecated::TabBarComponent < Deprecated::ApplicationV6Component
  def initialize(view_context, current_user:)
    super view_context
    @current_user = current_user
  end

  def render
    build_html do |h|
      h.tag :div, class: "c-tab-bar bg-white", data_controller: "tab-bar" do
        h.tag :div, class: "d-block h-100 navbar navbar-expand navbar-white px-0" do
          h.tag :ul, class: "align-items-center h-100 justify-content-around navbar-nav" do
            h.tag :li, class: "col nav-item px-0 text-center", data_action: "click->tab-bar#showSidebar" do
              h.tag :a, href: "#", class: "text-dark" do
                h.tag :i, class: "far fa-bars"

                h.tag :div, class: "small mt-1" do
                  h.text t("noun.menu")
                end
              end
            end

            h.tag :li, class: "col nav-item px-0 text-center" do
              h.tag :a, href: view_context.root_path, class: "text-dark" do
                h.tag :i, class: "far fa-home"

                h.tag :div, class: "small mt-1" do
                  h.text t("noun.home")
                end
              end
            end

            if @current_user
              h.tag :li, class: "col nav-item px-0 text-center" do
                h.tag :a, href: view_context.track_path, class: "text-dark" do
                  h.tag :i, class: "far fa-check"

                  h.tag :div, class: "small mt-1" do
                    h.text t("verb.track")
                  end
                end
              end

              h.tag :li, class: "col nav-item px-0 text-center" do
                h.tag :a, href: "/@#{@current_user.username}/watching", class: "text-dark" do
                  h.tag :i, class: "far fa-play"

                  h.tag :div, class: "small mt-1" do
                    h.text t("noun.library")
                  end
                end
              end

              h.tag :li, class: "col nav-item px-0 text-center" do
                h.tag :a, href: "/works/#{ENV.fetch("ANNICT_CURRENT_SEASON")}", class: "text-dark" do
                  h.tag :i, class: "far fa-#{Season.current.icon_name}"

                  h.tag :div, class: "small mt-1" do
                    h.text t("noun.current_season_shoten")
                  end
                end
              end
            else
              h.tag :li, class: "col nav-item px-0 text-center" do
                h.tag :a, href: "/works/#{ENV.fetch("ANNICT_CURRENT_SEASON")}", class: "text-dark" do
                  h.tag :i, class: "far fa-#{Season.current.icon_name}"

                  h.tag :div, class: "small mt-1" do
                    h.text t("noun.current_season_shoten")
                  end
                end
              end

              h.tag :li, class: "col nav-item px-0 text-center" do
                h.tag :a, href: view_context.sign_up_path, class: "text-dark" do
                  h.tag :i, class: "far fa-rocket"

                  h.tag :div, class: "small mt-1" do
                    h.text t("noun.sign_up_shorten")
                  end
                end
              end

              h.tag :li, class: "col nav-item px-0 text-center" do
                h.tag :a, href: view_context.sign_in_path, class: "text-dark" do
                  h.tag :i, class: "far fa-sign-in-alt"

                  h.tag :div, class: "small mt-1" do
                    h.text t("noun.sign_in")
                  end
                end
              end
            end
          end
        end
      end
    end
  end
end
