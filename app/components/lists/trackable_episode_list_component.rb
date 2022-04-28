# frozen_string_literal: true

module Lists
  class TrackableEpisodeListComponent < ApplicationV6Component
    def initialize(view_context, library_entries:, controller: nil, action: nil)
      super view_context
      @library_entries = library_entries
      @controller = controller
      @action = action
    end

    def render
      build_html do |h|
        h.tag :div, id: "trackable-episode-list" do
          if @library_entries.present?
            h.tag :div, class: "u-container-flat u-container-narrow" do
              h.tag :div, class: "gx-0 gy-3 row" do
                @library_entries.each do |le|
                  program = le.program
                  episode = le.next_episode
                  started_at = le.next_slot&.started_at

                  h.tag :div, class: "card u-card-flat" do
                    h.tag :div, class: "card-body" do
                      h.tag :div, class: "gx-3 row" do
                        h.tag :div, class: "col-auto" do
                          h.tag :div, {
                            class: "fw-bold u-cursor-pointer",
                            data_controller: "tracking-offcanvas-button",
                            data_tracking_offcanvas_button_frame_path: view_context.fragment_trackable_work_path(le.work_id),
                            data_action: "click->tracking-offcanvas-button#open"
                          } do
                            h.html view_context.render(Pictures::WorkPictureComponent.new(work: le.work, width: 80))
                          end
                        end

                        h.tag :div, class: "col" do
                          h.tag :div, {
                            class: "small u-cursor-pointer",
                            data_controller: "tracking-offcanvas-button",
                            data_tracking_offcanvas_button_frame_path: view_context.fragment_trackable_work_path(le.work_id),
                            data_action: "click->tracking-offcanvas-button#open"
                          } do
                            h.text le.work.local_title
                          end

                          if episode
                            h.tag :div, {
                              class: "fw-bold mt-1 u-cursor-pointer",
                              data_controller: "tracking-offcanvas-button",
                              data_tracking_offcanvas_button_frame_path: view_context.fragment_trackable_episode_path(episode.id),
                              data_action: "click->tracking-offcanvas-button#open"
                            } do
                              h.text episode.title_with_number
                            end
                          else
                            h.tag :div, class: "mt-1" do
                              h.tag :i, class: "far fa-check-circle me-1 text-success"
                              h.text t("messages.tracks.no_trackable_episodes")
                            end
                          end

                          if program
                            h.tag :div, class: "mt-1 small" do
                              if program.vod_title_url.present?
                                h.tag :a, href: program.vod_title_url, class: "text-muted", rel: "noopener", target: "_blank" do
                                  h.text program.channel.name
                                  h.tag :i, class: "fas fa-external-link-alt ps-1"
                                end
                              else
                                h.tag :span, class: "text-muted" do
                                  h.text program.channel.name
                                end
                              end
                            end
                          end

                          if started_at
                            h.tag :div, class: "small" do
                              h.tag :span, class: "text-muted" do
                                h.html display_time(started_at)
                              end

                              h.html Badges::SlotStartDateBadgeComponent.new(
                                view_context,
                                started_at: started_at,
                                time_zone: current_user.time_zone,
                                class_name: "ms-1"
                              ).render
                            end
                          end
                        end
                      end

                      if episode
                        h.tag :div, class: "align-items-center gx-3 mt-3 row" do
                          h.tag :div, class: "col-4" do
                            h.html Buttons::WatchEpisodeButtonComponent.new(
                              view_context,
                              episode_id: episode.id, button_text: t("verb.track"), class_name: "btn-sm btn-outline-info rounded-3 w-100"
                            ).render
                          end

                          h.tag :div, class: "col-4" do
                            h.tag :div, {
                              class: "btn btn-sm btn-outline-info rounded-3 w-100",
                              data_controller: "tracking-offcanvas-button",
                              data_tracking_offcanvas_button_frame_path: view_context.fragment_trackable_episode_path(episode.id),
                              data_action: "click->tracking-offcanvas-button#open"
                            } do
                              h.tag :i, class: "far fa-comment-check"

                              h.tag :span, class: "d-none d-sm-inline ms-1" do
                                h.text t("verb.track_with_comment")
                                h.text "..."
                              end

                              h.tag :span, class: "d-inline d-sm-none ms-1" do
                                h.text t("noun.comment_alt")
                                h.text "..."
                              end
                            end
                          end

                          h.tag :div, class: "col-4" do
                            h.html Buttons::SkipEpisodeButtonComponent.new(
                              view_context,
                              episode_id: episode.id, button_text: t("verb.skip_shorten"), class_name: "btn-sm btn-outline-secondary rounded-3 w-100"
                            ).render
                          end
                        end
                      end
                    end
                  end
                end
              end
            end

            h.tag :div, class: "mt-3 text-center" do
              h.html ButtonGroups::PaginationButtonGroupComponent.new(view_context, collection: @library_entries, controller: @controller, action: @action).render
            end
          else
            h.tag :div, class: "container u-container-flat" do
              h.tag :div, class: "card u-card-flat" do
                h.tag :div, class: "card-body" do
                  h.html EmptyV6Component.new(view_context, text: t("messages.tracks.no_trackable_episodes")).render
                end
              end
            end
          end
        end
      end
    end
  end
end
