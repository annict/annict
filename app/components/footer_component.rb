# frozen_string_literal: true

class FooterComponent < ApplicationComponent
  include ApplicationHelper

  def call
    Htmlrb.build do |el|
      el.div class: "c-footer py-5" do
        el.div class: "c-footer__main" do
          el.div class: "container-fluid" do
            el.div class: "row" do
              el.div class: "col-12 col-sm-3 mb-3 mb-sm-0" do
                el.h2 class: "mb-1" do
                  el.a href: "/" do
                    "Annict"
                  end
                end

                el.div class: "c-footer__description mb-3 small" do
                  "The platform for anime addicts."
                end

                el.div class: "c-footer__social-links" do
                  el.h4 do
                    el.div class: "row" do
                      [
                        ["https://twitter.com/#{helpers.twitter_username}", "twitter"],
                        ["https://github.com/annict", "github"],
                        [ENV.fetch("ANNICT_DISCORD_INVITE_URL"), "discord"]
                      ].each do |(url, icon_name)|
                        el.div class: "col col-auto" do
                          el.a(
                            class: "h2",
                            href: url,
                            target: "_blank",
                            rel: "noopener"
                          ) do
                            el.i(class: "fab fa-#{icon_name}") {}
                          end
                        end
                      end
                      nil
                    end
                  end
                end
              end

              el.div class: "col-12 col-sm-3 mb-3 mb-sm-0" do
                el.h4 class: "font-weight-bold mb-3" do
                  t "noun.services"
                end

                el.ul class: "c-footer__list list-unstyled" do
                  [
                    [t("noun.annict_userland"), userland_root_path],
                    [t("noun.annict_forum"), forum_root_path],
                    [t("noun.annict_db"), db_root_path],
                    [t("noun.annict_developers"), "https://developers.annict.jp"],
                    [t("noun.annict_supporters"), supporters_path]
                  ].each do |(link_title, link_url)|
                    el.li class: "mb-2" do
                      el.a href: link_url, rel: "noopener", target: "_blank" do
                        link_title
                      end
                    end
                  end
                  nil
                end
              end

              el.div class: "col-12 col-sm-3 mb-3 mb-sm-0" do
                el.h4 class: "font-weight-bold mb-3" do
                  t "noun.contents"
                end

                el.ul class: "c-footer__list list-unstyled" do
                  [
                    [t("head.title.faqs.index"), faqs_path, true],
                    [t("head.title.pages.about"), about_path, false],
                    [t("noun.terms_of_use"), terms_path, true],
                    [t("noun.privacy_policy"), privacy_path, true],
                    [t("head.title.pages.legal"), legal_path, true]
                  ].each do |(link_title, link_url, is_for_japanese)|
                    next if !helpers.locale_ja? && is_for_japanese

                    el.li class: "mb-2" do
                      el.a href: link_url, target: "_blank" do
                        link_title
                      end
                    end
                  end
                  nil
                end
              end

              el.div class: "col-12 col-sm-3" do
                el.h4 class: "font-weight-bold mb-3" do
                  t "noun.seasonal_anime"
                end

                el.ul class: "c-footer__list list-unstyled" do
                  Season.latest_slugs.each do |slug|
                    el.li class: "mb-2" do
                      year, name = slug.split("-")
                      el.a href: season_works_path(slug: slug) do
                        Season.new(year, name).local_name
                      end
                    end
                  end
                  nil
                end
              end
            end
          end
        end

        el.div class: "c-footer__auxiliary" do
          el.div class: "container-fluid py-2" do
            el.div class: "align-items-center row" do
              el.div class: "col-6" do
                el.h4 class: "d-inline-block font-weight-bold mb-0 mr-2 small" do
                  "#{t('noun.language')}:"
                end
                el.ul class: "d-inline-block list-inline mb-0" do
                  [
                    ["日本語", helpers.local_url_with_path(locale: :ja)],
                    ["English", helpers.local_url_with_path(locale: :en)]
                  ].each do |link_title, link_url|
                    el.li class: "list-inline-item" do
                      el.a href: link_url do
                        link_title
                      end
                    end
                  end
                  nil
                end
              end

              el.div class: "col-6 text-right" do
                el.div class: "c-footer__copyright small" do
                  el.i(class: "fal fa-copyright mr-1") {}
                  "2014-2020 Annict"
                end
              end
            end
          end
        end
      end
    end.html_safe
  end
end
