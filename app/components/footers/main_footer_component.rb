# frozen_string_literal: true

module Footers
  class MainFooterComponent < ApplicationV6Component
    def render
      build_html do |h|
        h.tag :div, class: "c-footer py-5" do
          h.tag :div, class: "c-footer__main container-fluid" do
            h.tag :div, class: "row" do
              h.tag :div, class: "col-12 col-sm-3 mb-3 mb-sm-0" do
                h.tag :h2, class: "mb-1" do
                  h.tag :a, class: "text-body", href: view_context.root_path do
                    h.text "Annict"
                  end
                end

                h.tag :div, class: "c-footer__description mb-3 small" do
                  h.text "A platform for anime addicts."
                end

                h.tag :div, class: "c-footer__social-links row" do
                  social_urls.each do |url, icon_name|
                    h.tag :div, class: "col col-auto" do
                      h.tag :a, aria_label: icon_name.humanize, href: url, class: "h2", target: "_blank", rel: "noopener" do
                        h.tag :i, class: "fab fa-#{icon_name}"
                      end
                    end
                  end
                end
              end

              h.tag :div, class: "col-12 col-sm-3 mb-3 mb-sm-0" do
                h.tag :h6, class: "fw-bold h4 mb-3" do
                  h.text t("noun.services")
                end

                h.tag :ul, class: "c-footer__list list-unstyled" do
                  service_urls.each do |link_url, link_title|
                    h.tag :li, class: "mb-2" do
                      h.tag :a, href: link_url, rel: "noopener", target: "_blank" do
                        h.text link_title
                      end
                    end
                  end
                end
              end

              h.tag :div, class: "col-12 col-sm-3 mb-3 mb-sm-0" do
                h.tag :h6, class: "fw-bold h4 mb-3" do
                  h.text t("noun.contents")
                end

                h.tag :ul, class: "c-footer__list list-unstyled" do
                  content_urls.each do |link_url, link_title, is_for_japanese|
                    next if !view_context.locale_ja? && is_for_japanese

                    h.tag :li, class: "mb-2" do
                      h.tag :a, href: link_url, target: "_blank" do
                        h.text link_title
                      end
                    end
                  end
                end
              end

              h.tag :div, class: "col-12 col-sm-3" do
                h.tag :h6, class: "fw-bold h4 mb-3" do
                  h.text t("noun.seasonal_work")
                end

                h.tag :ul, class: "c-footer__list list-unstyled" do
                  Season.latest_slugs.each do |slug|
                    h.tag :li, class: "mb-2" do
                      year, name = slug.split("-")
                      h.tag :a, href: view_context.seasonal_work_list_path(slug) do
                        h.text Season.new(year, name).local_name
                      end
                    end
                  end
                end
              end
            end
          end

          h.tag :div, class: "c-footer__auxiliary" do
            h.tag :div, class: "container-fluid py-2" do
              h.tag :div, class: "align-items-center row" do
                h.tag :div, class: "col-6" do
                  h.tag :h4, class: "d-inline-block fw-bold mb-0 me-2 small" do
                    h.text "#{t("noun.language")}:"
                  end

                  h.tag :ul, class: "d-inline-block list-inline mb-0" do
                    [
                      [view_context.local_url_with_path(locale: :ja), "日本語"],
                      [view_context.local_url_with_path(locale: :en), "English"]
                    ].each do |link_url, link_title|
                      h.tag :li, class: "list-inline-item" do
                        h.tag :a, href: link_url do
                          h.text link_title
                        end
                      end
                    end
                  end
                end

                h.tag :div, class: "col-6 text-end" do
                  h.tag :div, class: "c-footer__copyright small" do
                    h.tag :i, class: "fal fa-copyright me-1"
                    h.text "2014-2022 Annict"
                  end
                end
              end
            end
          end
        end
      end
    end

    private

    def social_urls
      [
        ["https://twitter.com/#{view_context.twitter_username}", "twitter"],
        ["https://github.com/annict", "github"],
        [ENV.fetch("ANNICT_DISCORD_INVITE_URL"), "discord"]
      ]
    end

    def service_urls
      [
        [view_context.userland_path, t("noun.annict_userland")],
        [view_context.forum_path, t("noun.annict_forum")],
        [view_context.db_root_path, t("noun.annict_db")],
        ["https://developers.annict.jp", t("noun.annict_developers")],
        [view_context.supporters_path, t("noun.annict_supporters")]
      ]
    end

    def content_urls
      [
        [view_context.faq_path, t("noun.faq"), true],
        [view_context.terms_path, t("noun.terms_of_use"), true],
        [view_context.privacy_path, t("noun.privacy_policy"), true],
        [view_context.legal_path, t("head.title.pages.legal"), true]
      ]
    end
  end
end
