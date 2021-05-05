# frozen_string_literal: true

module Sidebars
  class AnimeSidebarComponent < ApplicationComponent
    def initialize(view_context, anime:, vod_channels:)
      super view_context
      @anime = anime
      @vod_channels = vod_channels
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-anime-sidebar" do
          h.tag :div, class: "row mb-3 mb-sm-0" do
            h.tag :div, class: "col-4 col-sm-12" do
              h.tag :div, class: "mb-2 text-sm-left" do
                h.tag :a, href: view_context.anime_path(anime_id: @anime.id) do
                  h.html Images::AnimeImageComponent.new(
                    view_context,
                    image_url_1x: @anime.image_url(size: "350x"),
                    image_url_2x: @anime.image_url(size: "700x"),
                    alt: @anime.local_title
                  ).render
                end

                h.tag :div, class: "u-very-small text-muted" do
                  h.tag :i, class: "far fa-copyright"
                  h.text @anime.copyright
                end
              end
            end

            h.tag :div, class: "col-8 col-sm-12 pl-0 pl-sm-3" do
              h.tag :h1, class: "font-weight-bold h2 mb-3 text-sm-left" do
                h.tag :a, href: view_context.anime_path(anime_id: @anime.id), class: "u-text-body" do
                  h.text @anime.local_title
                end
              end

              h.tag :div, class: "row mb-3" do
                h.tag :div, class: "col text-center" do
                  h.tag :div, class: "h4 font-weight-bold mb-1" do
                    h.text @anime.watchers_count
                  end

                  h.tag :div, class: "text-muted small" do
                    h.text t("noun.watchers_count")
                  end
                end

                h.tag :div, class: "col text-center" do
                  h.tag :div, class: "h4 font-weight-bold mb-1" do
                    h.text @anime.satisfaction_rate.presence || "-"
                    h.tag :span, class: "small ml-1" do
                      h.text "%"
                    end
                  end

                  h.tag :div, class: "text-muted small" do
                    h.text t("noun.satisfaction_rate_shorten")
                  end
                end

                h.tag :div, class: "col text-center" do
                  h.tag :div, class: "h4 font-weight-bold mb-1" do
                    h.text @anime.ratings_count
                  end

                  h.tag :div, class: "text-muted small" do
                    h.text t("noun.ratings_count")
                  end
                end
              end

              h.tag :div, class: "mb-3" do
                h.html Selectors::StatusSelectorComponent.new(
                  view_context,
                  anime: @anime,
                  page_category: page_category
                ).render
              end
            end
          end

          h.tag :div, class: "mb-3" do
            h.tag :h2, class: "font-weight-bold h4 mb-3" do
              h.text t("noun.information")
            end

            h.tag :div, class: "mb-3 row" do
              if @anime.title_kana.present?
                h.tag :div, class: "col-6 col-sm-12 mb-2" do
                  h.tag :div, class: "small" do
                    h.text t("activerecord.attributes.work.title_kana")
                  end

                  h.tag :div do
                    h.text @anime.title_kana
                  end
                end
              end

              if @anime.title_alter.present?
                h.tag :div, class: "col-6 col-sm-12 mb-2" do
                  h.tag :div, class: "small" do
                    h.text t("activerecord.attributes.work.title_alter")
                  end

                  h.tag :div do
                    h.text @anime.title_alter
                  end
                end
              end

              if @anime.title_en.present?
                h.tag :div, class: "col-6 col-sm-12 mb-2" do
                  h.tag :div, class: "small" do
                    h.text t("activerecord.attributes.work.title_en")
                  end

                  h.tag :div do
                    h.text @anime.title_en
                  end
                end
              end

              if @anime.title_alter_en.present?
                h.tag :div, class: "col-6 col-sm-12 mb-2" do
                  h.tag :div, class: "small" do
                    h.text t("activerecord.attributes.work.title_alter_en")
                  end

                  h.tag :div do
                    h.text @anime.title_alter_en
                  end
                end
              end

              h.tag :div, class: "col-6 col-sm-12 mb-2" do
                h.tag :div, class: "small" do
                  h.text t("activerecord.attributes.work.media")
                end

                h.tag :div do
                  h.text @anime.media.text
                end
              end

              if @anime.season.present?
                h.tag :div, class: "col-6 col-sm-12 mb-2" do
                  h.tag :div, class: "small" do
                    h.text t("noun.release_season")
                  end

                  h.tag :div do
                    h.html @anime.release_season_link
                  end
                end
              end

              if @anime.started_on.present?
                h.tag :div, class: "col-6 col-sm-12 mb-2" do
                  h.tag :div, class: "small" do
                    h.text @anime.started_on_label
                  end

                  h.tag :div do
                    h.text display_date(@anime.started_on)
                  end
                end
              end

              if @anime.official_site_url.present?
                h.tag :div, class: "col-6 col-sm-12 mb-2" do
                  h.tag :div, class: "small" do
                    h.text t("activerecord.attributes.work.official_site_url")
                  end

                  h.tag :div do
                    h.html link_with_domain(@anime.official_site_url)
                  end
                end
              end

              if @anime.official_site_url_en.present?
                h.tag :div, class: "col-6 col-sm-12 mb-2" do
                  h.tag :div, class: "small" do
                    h.text t("activerecord.attributes.work.official_site_url_en")
                  end

                  h.tag :div do
                    h.html link_with_domain(@anime.official_site_url_en)
                  end
                end
              end

              if @anime.twitter_username.present?
                h.tag :div, class: "col-6 col-sm-12 mb-2" do
                  h.tag :div, class: "small" do
                    h.text t("activerecord.attributes.work.twitter_username")
                  end

                  h.tag :div do
                    h.html @anime.twitter_username_link
                  end
                end
              end

              if @anime.twitter_hashtag.present?
                h.tag :div, class: "col-6 col-sm-12 mb-2" do
                  h.tag :div, class: "small" do
                    h.text t("activerecord.attributes.work.twitter_hashtag")
                  end

                  h.tag :div do
                    h.html @anime.twitter_hashtag_link
                  end
                end
              end

              if @anime.wikipedia_url.present?
                h.tag :div, class: "col-6 col-sm-12 mb-2" do
                  h.tag :div, class: "small" do
                    h.text t("activerecord.attributes.work.wikipedia_url")
                  end

                  h.tag :div do
                    h.html link_with_domain(@anime.wikipedia_url)
                  end
                end
              end

              if @anime.wikipedia_url_en.present?
                h.tag :div, class: "col-6 col-sm-12 mb-2" do
                  h.tag :div, class: "small" do
                    h.text t("activerecord.attributes.work.wikipedia_url_en")
                  end

                  h.tag :div do
                    h.html link_with_domain(@anime.wikipedia_url_en)
                  end
                end
              end

              if @anime.syobocal_tid.present?
                h.tag :div, class: "col-6 col-sm-12 mb-2" do
                  h.tag :div, class: "small" do
                    h.text t("noun.syobocal")
                  end

                  h.tag :div do
                    h.html @anime.syobocal_link
                  end
                end
              end

              if @anime.mal_anime_id.present?
                h.tag :div, class: "col-6 col-sm-12 mb-2" do
                  h.tag :div, class: "small" do
                    h.text t("noun.my_anime_list")
                  end

                  h.tag :div do
                    h.html @anime.mal_anime_link
                  end
                end
              end
            end
          end

          if @vod_channels.present?
            h.tag :div, class: "mb-sm-1" do
              h.tag :h2, class: "font-weight-bold h4 mb-3" do
                h.text t("noun.vods")
              end

              h.tag :ul, class: "list-unstyled mb-0 row" do
                @vod_channels.each do |vod_channel|
                  h.tag :li, class: "col-6 col-sm-12 mb-3 mb-sm-2" do
                    h.tag :a, href: vod_channel.programs.first.vod_title_url, rel: "noopener", target: "_blank" do
                      h.text vod_channel.name
                    end
                  end
                end
              end
            end
          end

          h.tag :div, class: "mb-3" do
            h.tag :h2, class: "font-weight-bold h4 mb-3" do
              h.text t("noun.share")
            end

            h.html Buttons::ShareToTwitterButtonComponent.new(
              view_context,
              text: @anime.local_title,
              url: "#{local_url}/works/#{@anime.id}",
              hashtags: @anime.twitter_hashtag.presence || ""
            ).render

            h.html Buttons::ShareToFacebookButtonComponent.new(
              view_context,
              url: "#{local_url}/works/#{@anime.id}"
            ).render
          end

          h.tag :div,
            class: "mt-3",
            data_user_data_fetcher__anime_sidebar_target: "replacement",
            id: "c-anime-sidebar__db-link" do
              if current_user&.committer?
                h.tag :a, href: view_context.db_edit_work_path(@anime.id), class: "btn btn-secondary w-100 mt-2" do
                  h.text t("messages._common.edit_on_annict_db")
                end
              end
            end
        end
      end
    end
  end
end
