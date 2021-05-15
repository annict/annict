# frozen_string_literal: true

module Lists
  class TrackableEpisodeListComponent < ApplicationComponent
    def initialize(view_context, library_entries:, trackable_episodes:, slots:, page_category: "")
      super view_context
      @library_entries = library_entries
      @trackable_episodes = trackable_episodes
      @slots = slots
      @page_category = page_category
    end

    def render
      build_html do |h|
        h.tag :div, id: "trackable-episode-list" do
          h.tag :div, class: "card" do
            @library_entries.each do |le|
              h.tag :div, class: "card-header" do
                h.tag :div, class: "row" do
                  h.tag :div, class: "col" do
                    h.tag :div,
                      class: "u-cursor-pointer",
                      data_controller: "tracking-modal-button",
                      data_tracking_modal_button_frame_path: view_context.fragment_trackable_anime_path(le.work_id),
                      data_action: "click->tracking-modal-button#open" do
                        h.text le.work.local_title
                      end
                  end

                  h.tag :div, class: "col text-end" do
                    if le.program
                      if le.program.vod_title_url.present?
                        h.tag :a, href: le.program.vod_title_url, class: "text-body", rel: "noopener", target: "_blank" do
                          h.text le.program.channel.name
                          h.tag :i, class: "fas fa-external-link-alt pe-1"
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
                          episode_id: episode.id, page_category: @page_category, class_name: "btn-sm btn-outline-info rounded-circle").render
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
                        started_at = @slots.filter { |slot| slot.program_id == le.program.id && slot.episode_id == episode.id }.first&.started_at

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

          h.tag :div, class: "mt-3 text-center" do
            h.html paginate(@library_entries, params: {controller: "/tracks", action: "show"})
          end
        end
      end
    end
  end
end
