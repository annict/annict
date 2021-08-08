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
              h.tag :div, class: "col-12 col-sm-auto" do
                h.tag :div, class: "c-anime-header__anime-picture text-center" do
                  h.tag :a, href: view_context.anime_path(anime_id: @anime.id) do
                    h.html Pictures::AnimePictureComponent.new(view_context, anime: @anime, width: 180).render
                  end

                  h.tag :div, class: "mt-2 text-center" do
                    h.html ButtonGroups::AnimeButtonGroupComponent.new(view_context, anime: @anime).render
                  end
                end
              end

              h.tag :div, class: "col mt-3 mt-sm-0" do
                h.tag :div do
                  h.html Badges::AnimeMediaBadgeComponent.new(view_context, anime: @anime).render
                  h.html Badges::AnimeSeasonBadgeComponent.new(view_context, anime: @anime, class_name: "ms-2").render
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
                      h.text t("noun.ratings_count")
                      h.text ":"
                    end

                    h.tag :span, class: "fw-bold ms-1" do
                      h.text @anime.no_episodes? ? "-" : @anime.ratings_count
                    end
                  end
                end

                h.tag :ul, class: "list-inline mb-0" do
                  if locale_ja? && @anime.official_site_url.present?
                    h.tag :li, class: "list-inline-item" do
                      h.tag :a, href: @anime.official_site_url, target: "_blank", rel: "noopener" do
                        h.text t("activerecord.attributes.anime.official_site_url")
                        h.tag :i, class: "fas fa-external-link-alt ms-1 small"
                      end
                    end
                  elsif locale_en? && @anime.official_site_url_en.present?
                    h.tag :li, class: "list-inline-item" do
                      h.tag :a, href: @anime.official_site_url_en, target: "_blank", rel: "noopener" do
                        h.text t("activerecord.attributes.anime.official_site_url")
                        h.tag :i, class: "fas fa-external-link-alt ms-1 small"
                      end
                    end
                  end

                  if locale_ja? && @anime.wikipedia_url.present?
                    h.tag :li, class: "list-inline-item" do
                      h.tag :a, href: @anime.wikipedia_url, target: "_blank", rel: "noopener" do
                        h.text t("activerecord.attributes.anime.wikipedia_url")
                        h.tag :i, class: "fas fa-external-link-alt ms-1 small"
                      end
                    end
                  elsif locale_en? && @anime.wikipedia_url_en.present?
                    h.tag :li, class: "list-inline-item" do
                      h.tag :a, href: @anime.wikipedia_url_en, target: "_blank", rel: "noopener" do
                        h.text t("activerecord.attributes.anime.wikipedia_url")
                        h.tag :i, class: "fas fa-external-link-alt ms-1 small"
                      end
                    end
                  end
                end

                if @programs.present?
                  h.tag :ul, class: "list-inline mt-2" do
                    @programs.each do |program|
                      h.tag :li, class: "list-inline-item mt-2" do
                        h.tag :a, class: "btn btn-outline-primary btn-sm rounded-pill", href: program.vod_title_url, target: "_blank", rel: "noopener" do
                          h.text program.channel.name

                          if program.vod_title_name.present? && program.vod_title_name != @anime.title
                            h.tag :span, class: "ms-1" do
                              h.text "(#{program.vod_title_name})"
                            end
                          end

                          h.tag :i, class: "fas fa-external-link-alt ms-1 small"
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

                if current_user&.committer?
                  h.tag :div, class: "mt-3" do
                    h.tag :a, href: view_context.db_edit_work_path(@anime) do
                      h.text t("messages._common.edit_on_annict_db")
                    end
                  end
                end
              end
            end
          end

          h.tag :div, class: "c-nav mt-2" do
            h.tag :ul, class: "c-nav__list" do
              h.tag :li, class: "c-nav__item" do
                h.html active_link_to t("noun.top"), view_context.anime_path(anime_id: @anime.id),
                  active: page_category.in?(%w[anime]),
                  class: "c-nav__link",
                  class_active: "c-nav__link--active"
              end

              h.tag :li, class: "c-nav__item" do
                h.html active_link_to t("noun.information"), view_context.anime_info_path(@anime.id),
                  active: page_category.in?(%w[anime-info]),
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

              # TODO: @anime.casts_count > 0
              if true # standard:disable Lint/LiteralAsCondition
                h.tag :li, class: "c-nav__item" do
                  h.html active_link_to t("noun.casts"), view_context.cast_list_path(@anime.id),
                    active: page_category.in?(%w[cast-list]),
                    class: "c-nav__link",
                    class_active: "c-nav__link--active"
                end
              end

              # TODO: @anime.staffs_count > 0
              if true # standard:disable Lint/LiteralAsCondition
                h.tag :li, class: "c-nav__item" do
                  h.html active_link_to t("noun.staffs"), view_context.staff_list_path(@anime.id),
                    active: page_category.in?(%w[staff-list]),
                    class: "c-nav__link",
                    class_active: "c-nav__link--active"
                end
              end

              # TODO: @anime.videos_count > 0
              if true # standard:disable Lint/LiteralAsCondition
                h.tag :li, class: "c-nav__item" do
                  h.html active_link_to t("noun.videos"), view_context.video_list_path(@anime.id),
                    active: page_category.in?(%w[video-list]),
                    class: "c-nav__link",
                    class_active: "c-nav__link--active"
                end
              end

              # TODO: @anime.series_list_count > 0
              if true # standard:disable Lint/LiteralAsCondition
                h.tag :li, class: "c-nav__item" do
                  h.html active_link_to t("noun.related_works"), view_context.related_anime_list_path(@anime.id),
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
