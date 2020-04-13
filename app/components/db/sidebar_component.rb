# frozen_string_literal: true

module Db
  class SidebarComponent < ApplicationComponent
    include AssetsHelper

    def call
      Htmlrb.build do |el|
        el.div class: "c-db-sidebar" do
          el.a class: "c-db-sidebar__lockup my-3 px-3", href: db_root_path do
            el.span class: "c-db-sidebar__lockup__logo" do
              "Annict"
            end
            el.span class: "c-db-sidebar__lockup__brand" do
              "DB"
            end
          end

          el.form action: db_search_path, class: "px-1", method: "get" do
            el.div class: "form-group" do
              el.input(
                class: "form-control",
                name: "q",
                placeholder: t("messages._common.search_with_keywords"),
                type: "text"
              )
            end
          end

          el.ul class: "c-db-sidebar__menu nav navbar-nav px-3" do
            el.li do
              el.a class: "d-inline-block", href: db_activity_list_path do
                I18n.t("noun.activities")
              end
            end
            el.li do
              el.a class: "d-inline-block", href: db_series_list_path do
                I18n.t("noun.series")
              end
            end
            el.li do
              el.a class: "d-inline-block", href: db_works_path do
                I18n.t("noun.works")
              end
            end
            el.li do
              el.a class: "d-inline-block", href: db_people_path do
                I18n.t("noun.people")
              end
            end
            el.li do
              el.a class: "d-inline-block", href: db_organizations_path do
                I18n.t("noun.organizations")
              end
            end
            el.li do
              el.a class: "d-inline-block", href: db_characters_path do
                I18n.t("noun.characters")
              end
            end
            el.li do
              el.a class: "d-inline-block", href: db_channel_groups_path do
                I18n.t("noun.channel_groups")
              end
            end
            el.li do
              el.a class: "d-inline-block", href: db_channels_path do
                I18n.t("noun.channels")
              end
            end
          end

          el.a class: "c-db-sidebar__annict-link d-inline-block text-center", href: root_path do
            el.img height: "40", src: asset_bundle_url("images/logos/color-white.png"), width: "33"
          end
        end
      end.html_safe
    end
  end
end
