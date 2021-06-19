# frozen_string_literal: true

module Headers
  class AnimeHeaderComponent < ApplicationV6Component
    def initialize(view_context, anime:, programs:)
      super view_context
      @anime = anime
      @programs = programs
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-anime-header pt-3" do
          h.tag :div, class: "container" do
            h.tag :div, class: "gx-3 row" do
              h.tag :div, class: "col-auto" do
                h.tag :div, class: "c-anime-header__anime-picture" do
                  h.tag :a, href: view_context.anime_path(anime_id: @anime.id) do
                    h.html Pictures::AnimePictureComponent.new(view_context, anime: @anime, width: 180).render
                  end
                end
              end

              h.tag :div, class: "col" do
                h.tag :div do
                  h.html Badges::AnimeMediaBadgeComponent.new(self, anime: @anime).render

                  if @anime.release_season_link.present?
                    h.tag :span, class: "ms-2" do
                      h.html @anime.release_season_link
                    end
                  end
                end

                h.tag :h1, class: "fw-bold h2 mt-1" do
                  h.tag :a, href: view_context.anime_path(anime_id: @anime.id), class: "text-body" do
                    h.text @anime.local_title
                  end
                end

                h.tag :ul, class: "list-inline" do
                  h.tag :li, class: "list-inline-item" do
                    h.tag :span, class: "small text-muted" do
                      h.text t("noun.watchers_count")
                      h.text ":"
                    end

                    h.tag :span, class: "fw-bold ms-1" do
                      h.text @anime.watchers_count
                    end
                  end

                  h.tag :li, class: "list-inline-item" do
                    h.tag :span, class: "small text-muted" do
                      h.text t("noun.satisfaction_rate_shorten")
                      h.text ":"
                    end

                    h.tag :span, class: "fw-bold ms-1" do
                      h.text @anime.satisfaction_rate.presence || "-"

                      h.tag :span, class: "ms-1 small" do
                        h.text "%"
                      end
                    end
                  end

                  h.tag :li, class: "list-inline-item" do
                    h.tag :span, class: "small text-muted" do
                      h.text t("noun.ratings_count")
                      h.text ":"
                    end

                    h.tag :span, class: "fw-bold ms-1" do
                      h.text @anime.no_episodes? ? "-" : @anime.ratings_count
                    end
                  end
                end

                h.tag :ul, class: "list-inline" do
                  if locale_ja? && @anime.official_site_url.present?
                    h.tag :li, class: "list-inline-item" do
                      h.tag :a, href: @anime.official_site_url, target: "_blank", rel: "noopener" do
                        h.text t("activerecord.attributes.work.official_site_url")
                      end
                    end
                  elsif locale_en? && @anime.official_site_url_en.present?
                    h.tag :li, class: "list-inline-item" do
                      h.tag :a, href: @anime.official_site_url_en, target: "_blank", rel: "noopener" do
                        h.text t("activerecord.attributes.work.official_site_url")
                      end
                    end
                  end

                  if locale_ja? && @anime.wikipedia_url.present?
                    h.tag :li, class: "list-inline-item" do
                      h.tag :a, href: @anime.wikipedia_url, target: "_blank", rel: "noopener" do
                        h.text t("activerecord.attributes.work.wikipedia_url")
                      end
                    end
                  elsif locale_en? && @anime.wikipedia_url_en.present?
                    h.tag :li, class: "list-inline-item" do
                      h.tag :a, href: @anime.wikipedia_url_en, target: "_blank", rel: "noopener" do
                        h.text t("activerecord.attributes.work.wikipedia_url")
                      end
                    end
                  end
                end

                if @programs.present?
                  h.tag :ul, class: "list-inline" do
                    @programs.each do |program|
                      h.tag :li, class: "list-inline-item" do
                        h.tag :a, class: "btn btn-outline-primary btn-sm", href: program.vod_title_url, target: "_blank", rel: "noopener" do
                          h.text program.channel.name

                          if program.vod_title_name.present? && program.vod_title_name != @anime.title
                            h.tag :span, class: "ms-1" do
                              h.text "(#{program.vod_title_name})"
                            end
                          end
                        end
                      end
                    end
                  end
                end

                if @anime.copyright
                  h.tag :div, class: "text-muted u-very-small" do
                    h.tag :i, class: "far fa-copyright me-1"
                    h.text @anime.copyright
                  end
                end
              end
            end
          end

          h.tag :div, class: "c-nav mt-2" do
            h.tag :ul, class: "c-nav__list" do
              h.tag :li, class: "c-nav__item" do
                h.html active_link_to t("noun.overview"), view_context.anime_path(anime_id: @anime.id),
                  active: page_category.in?(%w[anime]),
                  class: "c-nav__link",
                  class_active: "c-nav__link--active"
              end

              h.tag :li, class: "c-nav__item" do
                h.html active_link_to t("noun.information"), "#",
                  active: page_category.in?(%w[]),
                  class: "c-nav__link",
                  class_active: "c-nav__link--active"
              end

              h.tag :li, class: "c-nav__item" do
                h.html active_link_to t("noun.records"), view_context.anime_record_list_path(anime_id: @anime.id),
                  active: page_category.in?(%w[anime-record-list]),
                  class: "c-nav__link",
                  class_active: "c-nav__link--active"
              end

              unless @anime.no_episodes?
                h.tag :li, class: "c-nav__item" do
                  h.html active_link_to t("noun.episodes"), view_context.episode_list_path(anime_id: @anime.id),
                    active: page_category.in?(%w[episode episode-list]),
                    class: "c-nav__link",
                    class_active: "c-nav__link--active"
                end
              end

              if true # TODO: @anime.casts_count > 0
                h.tag :li, class: "c-nav__item" do
                  h.html active_link_to t("noun.casts"), "#",
                    active: page_category.in?(%w[cast-list]),
                    class: "c-nav__link",
                    class_active: "c-nav__link--active"
                end
              end

              if true # TODO: @anime.staffs_count > 0
                h.tag :li, class: "c-nav__item" do
                  h.html active_link_to t("noun.staffs"), "#",
                    active: page_category.in?(%w[staff-list]),
                    class: "c-nav__link",
                    class_active: "c-nav__link--active"
                end
              end

              if true # TODO: @anime.videos_count > 0
                h.tag :li, class: "c-nav__item" do
                  h.html active_link_to t("noun.videos"), "#",
                    active: page_category.in?(%w[video-list]),
                    class: "c-nav__link",
                    class_active: "c-nav__link--active"
                end
              end

              if true # TODO: @anime.series_list_count > 0
                h.tag :li, class: "c-nav__item" do
                  h.html active_link_to t("noun.related_works"), "#",
                    active: page_category.in?(%w[related-anime-list]),
                    class: "c-nav__link",
                    class_active: "c-nav__link--active"
                end
              end
            end
          end
        end
      end
    end
  end
end
