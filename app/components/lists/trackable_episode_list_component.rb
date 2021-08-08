# frozen_string_literal: true

module Lists
  class TrackableEpisodeListComponent < ApplicationV6Component
    def initialize(view_context, library_entries:, trackable_episodes:, slots:, controller: nil, action: nil)
      super view_context
      @library_entries = library_entries
      @trackable_episodes = trackable_episodes
      @slots = slots
      @controller = controller
      @action = action
    end

    def render
      build_html do |h|
        h.tag :div, id: "trackable-episode-list" do
          if @library_entries.present?
            h.tag :div, class: "container u-container-flat" do
              h.tag :div, class: "card u-card-flat" do
                @library_entries.each do |le|
                  h.tag :div, class: "card-header" do
                    h.tag :div, class: "row" do
                      h.tag :div, class: "col" do
                        h.tag :div,
                          class: "fw-bold u-cursor-pointer",
                          data_controller: "tracking-modal-button",
                          data_tracking_modal_button_frame_path: view_context.fragment_trackable_anime_path(le.work_id),
                          data_action: "click->tracking-modal-button#open" do
                            h.text le.anime.local_title
                          end
                      end

                      h.tag :div, class: "col text-end" do
                        if le.program
                          if le.program.vod_title_url.present?
                            h.tag :a, href: le.program.vod_title_url, class: "text-body", rel: "noopener", target: "_blank" do
                              h.text le.program.channel.name
                              h.tag :i, class: "fas fa-external-link-alt ps-1"
                            end
                          else
                            h.text le.program.channel.name
                          end
                        end
                      end
                    end
                  end

                  h.tag :ul, class: "list-group list-group-flush" do
                    @trackable_episodes.filter { |episode| episode.work_id == le.work_id }.each do |episode|
                      h.tag :li, class: "list-group-item" do
                        h.tag :div, class: "align-items-center row" do
                          h.tag :div, class: "col-auto pe-0" do
                            h.html Buttons::WatchEpisodeButtonComponent.new(view_context,
                              episode_id: episode.id, class_name: "btn-sm btn-outline-info rounded-circle").render
                          end

                          h.tag :div, class: "col-auto ps-2 pe-0" do
                            h.html Buttons::SkipEpisodeButtonComponent.new(view_context,
                              episode_id: episode.id, class_name: "btn-sm btn-outline-secondary rounded-circle").render
                          end

                          h.tag :div, class: "col" do
                            h.tag :div,
                              class: "u-cursor-pointer",
                              data_controller: "tracking-modal-button",
                              data_tracking_modal_button_frame_path: view_context.fragment_trackable_episode_path(episode.id),
                              data_action: "click->tracking-modal-button#open" do
                                h.text episode.title_with_number
                              end
                          end

                          if le.program
                            started_at = @slots.find { |slot| slot.program_id == le.program.id && slot.episode_id == episode.id }&.started_at

                            if started_at
                              h.tag :div, class: "col text-end" do
                                h.tag :span, class: "small text-muted" do
                                  h.html display_time(started_at)
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
