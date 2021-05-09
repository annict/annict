# frozen_string_literal: true

module Activities
  class RecordActivityComponent < ApplicationComponent
    def initialize(view_context, activity_group_struct:, page_category: "")
      super view_context
      @activity_group_struct = activity_group_struct
      @page_category = page_category
      @user = activity_group_struct.user.decorate
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-timeline__record-activity" do
          h.tag :div, class: "row g-3" do
            h.tag :div, class: "col-auto" do
              h.tag :a, href: view_context.profile_path(@user.username) do
                h.html Pictures::AvatarPictureComponent.new(view_context,
                  user: @user,
                  width: 32,
                  mb_width: 32
                ).render
              end
            end

            h.tag :div, class: "col" do
              h.tag :div do
                h.tag :span, class: "c-timeline__user-name" do
                  h.tag :a, href: view_context.profile_path(@user.username), class: "text-body fw-bold u-link" do
                    h.text @user.name
                  end
                end

                h.tag :small, class: "ms-1" do
                  h.text t("messages._components.activities.episode_record.created")
                end

                h.html RelativeTimeComponent.new(
                  view_context,
                  time: @activity_group_struct.created_at.iso8601,
                  class_name: "ms-1 small text-muted"
                ).render
              end

              if @activity_group_struct.outstanding
                record = @activity_group_struct.itemables.first
                if record.episode_record?
                  h.html Contents::EpisodeRecordContentComponent.new(view_context, record: record).render
                else
                  h.html Contents::AnimeRecordContentComponent.new(view_context, record: record).render
                end
              else
                h.tag :div, class: "c-timeline__activity-cards" do
                  @activity_group_struct.itemables.each do |record|
                    h.tag :div, class: "mb-3" do
                      if record.episode_record?
                        h.html Cards::EpisodeRecordCardComponent.new(view_context, episode_record: record.episode_record).render
                      else
                        h.html Cards::AnimeRecordCardComponent.new(view_context, anime_record: record.anime_record).render
                      end

                      h.html Footers::RecordFooterComponent.new(view_context, record: record).render
                    end
                  end
                end
              end

              if @activity_group_struct.itemables.length > 2
                h.tag :div, {
                  class: "text-center",
                  data_action: "click->timeline-activity#next",
                  data_target: "timeline-activity.nextButton"
                } do
                  h.tag :div, class: "text-center" do
                    h.tag :div, class: "c-activity-more-button btn btn-outline-secondary btn-small py-1" do
                      h.tag :i, class: "fal fa-chevron-double-down"
                      h.text t("messages._components.activities.episode_record.more", n: @activity_group_struct.itemables.length - 2)
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
