# frozen_string_literal: true

class SidebarComponent < ApplicationComponent
  def initialize(user:)
    @user = user
  end

  def call
    Htmlrb.build do |el|
      el.c_sidebar class: "h-100", "v-slot:default": "slotProps" do
        el.div(
          ":class": "{ active: slotProps.isSidebarOpen }",
          class: "c-sidebar__background",
          "v-on:click": "slotProps.hideSidebar"
        ) {}

        el.transition name: "slide" do
          el.div(
            ":class": "{ active: slotProps.isSidebarOpen }",
            class: "c-sidebar__content",
            v_show: "slotProps.isSidebarOpen"
          ) do
            el.a class: "c-sidebar__logo d-inline-block mb-3 py-3 text-center u-bg-mizuho w-100", href: "/" do
              el.img(
                alt: "Annict",
                class: "js-lazy",
                data_src: helpers.asset_bundle_url("images/logos/color-white.png"),
                height: "30",
                src: "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 1 1'%3E%3C/svg%3E",
                width: "25"
              )
            end

            if user
              el.div class: "mb-3 px-3" do
                el.a class: "align-content-center d-flex", href: user_path(user.username) do
                  el.img(
                    alt: user.username,
                    class: "rounded-circle",
                    height: "50",
                    src: helpers.ann_image_url(user.profile, :image, size: "50x50"),
                    width: "50"
                  )
                  el.div class: "ml-2" do
                    el.div class: "font-weight-bold u-text-body" do
                      user.profile.name
                    end
                    el.div class: "small text-secondary" do
                      "@#{user.username}"
                    end
                  end
                end
              end
            end

            el.div class: "mb-3 px-1" do
              el.form action: search_path, autocomplete: "off", class: "", method: "get" do
                el.input class: "form-control", name: "q", placeholder: t("messages._common.search_with_keywords"), type: "text"
              end
            end

            el.ul class: "list-unstyled px-3" do
              el.li do
                el.a class: "d-flex justify-content-between py-2 w-100", href: "/" do
                  el.div class: "c-sidebar__icon text-muted" do
                    el.i(class: "fal fa-home") {}
                  end
                  el.div class: "flex-grow-1 text-body" do
                    t "noun.home"
                  end
                end
              end

              if user
                el.li do
                  el.a class: "d-flex justify-content-between py-2 w-100", href: user_path(user.username) do
                    el.div class: "c-sidebar__icon text-muted" do
                      el.i(class: "fal fa-user") {}
                    end
                    el.div class: "flex-grow-1 text-body" do
                      t "noun.profile"
                    end
                  end
                end

                el.li do
                  el.a class: "d-flex justify-content-between py-2", href: notifications_path do
                    el.div class: "c-sidebar__icon text-muted" do
                      el.i(class: "fal fa-bell") {}
                    end

                    el.div class: "flex-grow-1 text-body" do
                      t "head.title.notifications.index"
                    end

                    if user.notifications_count > 0
                      el.div do
                        el.span class: "badge badge-pill badge-primary" do
                          user.notifications_count.to_s
                        end
                      end
                    end
                  end
                end

                el.li do
                  el.a class: "d-flex justify-content-between py-2", href: programs_path do
                    el.div class: "c-sidebar__icon text-muted" do
                      el.i(class: "fal fa-calendar") {}
                    end

                    el.div class: "flex-grow-1 text-body" do
                      t "noun.slots"
                    end
                  end
                end
              end
            end

            if user
              el.div class: "font-weight-bold mb-3 px-3 small text-secondary" do
                t "noun.library"
              end

              el.ul class: "list-unstyled px-3" do
                [
                  [:watching, "play", t("noun.watching")],
                  [:wanna_watch, "circle", t("noun.plan_to_watch")],
                  [:watched, "check", t("noun.completed")],
                  [:on_hold, "pause", t("noun.on_hold")],
                  [:stop_watching, "stop", t("noun.dropped")]
                ].each do |(status_kind, icon_name, link_text)|
                  el.li do
                    el.a(
                      class: "d-flex justify-content-between py-2",
                      href: library_path(username: user.username, status_kind: status_kind)
                    ) do
                      el.div class: "c-sidebar__icon text-muted" do
                        el.i(class: "fal fa-#{icon_name}") {}
                      end

                      el.div class: "flex-grow-1 text-body" do
                        link_text
                      end
                    end
                  end
                end
                nil
              end
            end

            el.div class: "font-weight-bold mb-3 px-3 small text-secondary" do
              t "verb.explore"
            end

            el.ul class: "list-unstyled px-3" do
              [
                [ENV.fetch("ANNICT_CURRENT_SEASON"), "snowflake", t("noun.current_season")],
                [ENV.fetch("ANNICT_NEXT_SEASON"), "flower", t("noun.next_season")],
                [ENV.fetch("ANNICT_PREVIOUS_SEASON"), "pumpkin", t("noun.previous_season")],
                [:popular, "fire", t("head.title.works.popular")],
                [:newest, "bolt", t("head.title.works.newest")]
              ].each do |(page_type, icon_name, link_text)|
                el.li do
                  el.a class: "d-flex justify-content-between py-2", href: "/works/#{page_type}" do
                    el.div class: "c-sidebar__icon text-muted" do
                      el.i(class: "fal fa-#{icon_name}") {}
                    end

                    el.div class: "flex-grow-1 text-body" do
                      link_text
                    end
                  end
                end
              end
              nil
            end

            el.div class: "font-weight-bold mb-3 px-3 small text-secondary" do
              "Misc"
            end

            el.ul class: "list-unstyled px-3" do
              [
                [friends_path, "search", t("head.title.friends.index"), true],
                [channels_path, "tv-retro", t("head.title.channels.index"), true],
                [profile_path, "cog", t("noun.settings"), true],
                [faqs_path, "question-circle", t("head.title.faqs.index"), false],
                [about_path, "info-circle", t("head.title.pages.about"), false]
              ].each do |(link_path, icon_name, link_text, sign_in_required)|
                if user || !sign_in_required
                  el.li do
                    el.a class: "d-flex justify-content-between py-2", href: link_path do
                      el.div class: "c-sidebar__icon text-muted" do
                        el.i(class: "fal fa-#{icon_name}") {}
                      end

                      el.div class: "flex-grow-1 text-body" do
                        link_text
                      end
                    end
                  end
                end
              end
              nil
            end

            el.div class: "font-weight-bold mb-3 px-3 small text-secondary" do
              t "noun.services"
            end

            el.ul class: "list-unstyled px-3" do
              [
                [userland_root_path, "signal-stream", t("noun.annict_userland")],
                [forum_root_path, "comments-alt", t("noun.annict_forum")],
                [db_root_path, "database", t("noun.annict_db")],
                [supporters_path, "sparkles", t("noun.annict_supporters")]
              ].each do |(link_path, icon_name, link_text)|
                el.li do
                  el.a class: "d-flex justify-content-between py-2", href: link_path do
                    el.div class: "c-sidebar__icon text-muted" do
                      el.i(class: "fal fa-#{icon_name}") {}
                    end

                    el.div class: "flex-grow-1 text-body" do
                      link_text
                    end
                  end
                end
              end

              el.li do
                el.a(
                  class: "d-flex justify-content-between py-2",
                  href: "https://developers.annict.jp",
                  rel: "noopener",
                  target: "_blank"
                ) do
                  el.div class: "c-sidebar__icon text-muted" do
                    el.i(class: "fal fa-code") {}
                  end

                  el.div class: "flex-grow-1 text-body" do
                    t "noun.annict_developers"
                  end
                end
              end
            end

            el.hr

            el.ul class: "list-unstyled px-3" do
              el.li do
                if user
                  el.a(
                    class: "d-flex justify-content-between py-2",
                    data_confirm: t("messages._common.are_you_sure"),
                    data_method: "delete",
                    href: destroy_user_session_path
                  ) do
                    el.div class: "c-sidebar__icon text-muted" do
                      el.i(class: "fal fa-sign-out") {}
                    end

                    el.div class: "flex-grow-1 text-body" do
                      t "verb.sign_out"
                    end
                  end
                else
                  el.a class: "d-flex justify-content-between py-2", href: new_user_session_path do
                    el.div class: "c-sidebar__icon text-muted" do
                      el.i(class: "fal fa-sign-in") {}
                    end

                    el.div class: "flex-grow-1 text-body" do
                      t "noun.sign_in"
                    end
                  end
                end
              end
            end
          end
        end
      end
    end.html_safe
  end

  private

  attr_reader :user
end
