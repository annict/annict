# frozen_string_literal: true

module Sidebars
  class MainSidebarComponent < ApplicationComponent
    def initialize(view_context, search:)
      super view_context
      @search = search
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-sidebar h-100" do
          h.tag :div, class: "c-sidebar__background", data_action: "click->sidebar#hide"

          h.tag :div, class: "c-sidebar__content" do
            h.tag :a, href: view_context.root_path, class: "c-sidebar__logo d-inline-block mb-3 text-center u-bg-mizuho w-100" do
              h.html image_tag(
                "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 1 1'%3E%3C/svg%3E",
                alt: "Annict",
                class: "js-lazy",
                data: {src: view_context.asset_bundle_url("images/logos/color-white.png")},
                size: "25x30"
              )
            end

            h.tag :div, class: "mb-3 px-3" do
              if current_user
                h.tag :a, href: view_context.profile_path(current_user.username), class: "align-content-center row" do
                  h.tag :div, class: "col-3 pr-0" do
                    h.html ProfileImageComponent.new(view_context, image_url_1x: ann_image_url(current_user.profile, :image, size: "50x50", ratio: "1:1"), alt: "@#{current_user.username}", lazy_load: false).render
                  end

                  h.tag :div, class: "col-auto" do
                    h.tag :div, class: "font-weight-bold u-text-body" do
                      h.text current_user.profile.name
                    end

                    h.tag :div, class: "small text-secondary" do
                      h.text "@#{current_user.username}"
                    end
                  end
                end
              else
                h.tag :a, href: view_context.sign_in_path, class: "d-flex justify-content-between py-2" do
                  h.tag :div, class: "c-sidebar__icon mr-1 text-center text-muted" do
                    h.tag :i, class: "far fa-sign-in"
                  end

                  h.tag :div, class: "flex-grow-1 text-body" do
                    h.text t("noun.sign_in")
                  end
                end
              end
            end

            h.tag :div, class: "mb-3 px-1" do
              h.tag :form, autocomplete: "off", method: :get, action: view_context.search_path do
                h.html text_field_tag("q", @search.q, class: "form-control", placeholder: t("messages._common.search_with_keywords"))
              end
            end

            h.tag :ul, class: "list-unstyled px-3" do
              h.tag :li do
                h.tag :a, href: view_context.root_path, class: "d-flex justify-content-between py-2" do
                  h.tag :div, class: "c-sidebar__icon mr-1 text-center text-muted" do
                    h.tag :i, class: "far fa-home"
                  end

                  h.tag :div, class: "flex-grow-1 text-body" do
                    h.text t("noun.home")
                  end
                end
              end

              h.tag :li do
                h.tag :a, href: current_user ? view_context.profile_path(current_user.username) : view_context.sign_in_path, class: "d-flex justify-content-between py-2" do
                  h.tag :div, class: "c-sidebar__icon mr-1 text-center text-muted" do
                    h.tag :i, class: "far fa-user"
                  end

                  h.tag :div, class: "flex-grow-1 text-body" do
                    h.text t("noun.profile")
                  end
                end
              end

              h.tag :li do
                h.tag :a, href: view_context.notifications_path, class: "d-flex justify-content-between py-2" do
                  h.tag :div, class: "c-sidebar__icon mr-1 text-center text-muted" do
                    h.tag :i, class: "far fa-bell"
                  end

                  h.tag :div, class: "flex-grow-1 text-body" do
                    h.text t("head.title.notifications.index")
                  end

                  h.tag :div do
                    if current_user && current_user.notifications_count > 0
                      h.tag :span, class: "badge badge-pill badge-primary" do
                        h.text current_user.notifications_count
                      end
                    end
                  end
                end
              end

              h.tag :li do
                h.tag :a, href: "/track", class: "d-flex justify-content-between py-2" do
                  h.tag :div, class: "c-sidebar__icon mr-1 text-center text-muted" do
                    h.tag :i, class: "far fa-tasks"
                  end

                  h.tag :div, class: "flex-grow-1 text-body" do
                    h.text t("verb.track")
                  end
                end
              end
            end

            h.tag :div, class: "font-weight-bold mb-3 px-3 small text-secondary" do
              h.text t("noun.library")
            end

            h.tag :ul, class: "list-unstyled px-3" do
              [
                [:watching, "play", t("noun.watching")],
                [:wanna_watch, "circle", t("noun.plan_to_watch")],
                [:watched, "check", t("noun.completed")],
                [:on_hold, "pause", t("noun.on_hold")],
                [:stop_watching, "stop", t("noun.dropped")]
              ].each do |status_kind, icon_name, link_text|
                h.tag :li do
                  h.tag :a, href: current_user ? view_context.library_path(username: current_user.username, status_kind: status_kind) : view_context.sign_in_path, class: "d-flex justify-content-between py-2" do
                    h.tag :div, class: "c-sidebar__icon mr-1 text-center text-muted" do
                      h.tag :i, class: "far fa-#{icon_name}"
                    end

                    h.tag :div, class: "flex-grow-1 text-body" do
                      h.text link_text
                    end
                  end
                end
              end
            end

            h.tag :div, class: "font-weight-bold mb-3 px-3 small text-secondary" do
              h.text t("verb.explore")
            end

            h.tag :ul, class: "list-unstyled px-3" do
              [
                [ENV.fetch("ANNICT_CURRENT_SEASON"), Season.current.icon_name, t("noun.current_season")],
                [ENV.fetch("ANNICT_NEXT_SEASON"), Season.next.icon_name, t("noun.next_season")],
                [ENV.fetch("ANNICT_PREVIOUS_SEASON"), Season.prev.icon_name, t("noun.previous_season")],
                [:popular, "fire-alt", t("head.title.works.popular")],
                [:newest, "bolt", t("head.title.works.newest")]
              ].each do |page_type, icon_name, link_text|
                h.tag :li do
                  h.tag :a, href: "/works/#{page_type}", class: "d-flex justify-content-between py-2" do
                    h.tag :div, class: "c-sidebar__icon mr-1 text-center text-muted" do
                      h.tag :i, class: "far fa-#{icon_name}"
                    end

                    h.tag :div, class: "flex-grow-1 text-body" do
                      h.text link_text
                    end
                  end
                end
              end
            end

            h.tag :div, class: "font-weight-bold mb-3 px-3 small text-secondary" do
              h.text "Misc"
            end

            h.tag :ul, class: "list-unstyled px-3" do
              [
                [view_context.friends_path, "search", t("head.title.friends.index")],
                [view_context.channels_path, "tv-retro", t("head.title.channels.index")],
                [view_context.profile_setting_path, "cog", t("noun.settings")],
                [view_context.faqs_path, "question-circle", t("head.title.faqs.index")],
                [view_context.about_path, "info-circle", t("head.title.pages.about")]
              ].each do |link_path, icon_name, link_text|
                h.tag :li do
                  h.tag :a, href: link_path, class: "d-flex justify-content-between py-2" do
                    h.tag :div, class: "c-sidebar__icon mr-1 text-center text-muted" do
                      h.tag :i, class: "far fa-#{icon_name}"
                    end

                    h.tag :div, class: "flex-grow-1 text-body" do
                      h.text link_text
                    end
                  end
                end
              end
            end

            h.tag :div, class: "font-weight-bold mb-3 px-3 small text-secondary" do
              h.text t("noun.services")
            end

            h.tag :ul, class: "list-unstyled px-3" do
              [
                [view_context.userland_root_path, "signal-stream", t("noun.annict_userland")],
                [view_context.forum_root_path, "comments-alt", t("noun.annict_forum")],
                [view_context.db_root_path, "database", t("noun.annict_db")],
                [view_context.supporters_path, "sparkles", t("noun.annict_supporters")]
              ].each do |link_path, icon_name, link_text|
                h.tag :li do
                  h.tag :a, href: link_path, class: "d-flex justify-content-between py-2" do
                    h.tag :div, class: "c-sidebar__icon mr-1 text-center text-muted" do
                      h.tag :i, class: "far fa-#{icon_name}"
                    end

                    h.tag :div, class: "flex-grow-1 text-body" do
                      h.text link_text
                    end
                  end
                end
              end

              h.tag :li do
                h.tag :a, href: "https://developers.annict.jp", class: "d-flex justify-content-between py-2", rel: "noopener", target: "_blank" do
                  h.tag :div, class: "c-sidebar__icon mr-1 text-center text-muted" do
                    h.tag :i, class: "far fa-code"
                  end

                  h.tag :div, class: "flex-grow-1 text-body" do
                    h.text t("noun.annict_developers")
                  end
                end
              end
            end

            if current_user
              h.tag :hr

              h.tag :ul, class: "list-unstyled px-3" do
                h.tag :li do
                  h.tag :a,
                    href: view_context.sign_out_path,
                    class: "d-flex justify-content-between py-2",
                    data_confirm: t("messages._common.are_you_sure"),
                    data_method: :delete do
                      h.tag :div, class: "c-sidebar__icon mr-1 text-center text-muted" do
                        h.tag :i, class: "far fa-sign-out"
                      end

                      h.tag :div, class: "flex-grow-1 text-body" do
                        h.text t("verb.sign_out")
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
end
