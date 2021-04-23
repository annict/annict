# frozen_string_literal: true

class RecordHeaderComponent2 < ApplicationComponent2
  def initialize(view_context, record:, current_user: nil)
    super view_context
    @record = record
    @current_user = current_user
  end

  def render
    build_html do |h|
      h.tag :div, class: "c-record-header row" do
        h.tag :div, class: "col-auto pr-0" do
          link_to view_context.profile_path(@record.user.username) do
            h.html ProfileImageComponent2.new(view_context,
              image_url_1x: @record.user.avatar_url(size: "50x50"),
              alt: "@#{@record.user.username}",
              lazy_load: false
            ).render
          end
        end

        h.tag :div, class: "col" do
          h.tag :div do
            h.tag :a, href: view_context.profile_path(@record.user.username), class: "font-weight-bold mr-1 text-body" do
              h.tag :span, class: "mr-1" do
                h.text @record.user.name
              end

              h.tag :small, class: "text-muted" do
                h.text "@#{@record.user.username}"
              end
            end

            if @record.user.supporter? && !@record.user.setting.hide_supporter_badge?
              h.tag :span, class: "badge u-badge-supporter" do
                h.text t("noun.supporter")
              end
            end
          end

          h.tag :div do
            h.tag :a, href: view_context.record_path(@record.user.username, @record.id), class: "small text-muted" do
              h.text display_time(@record.created_at)
            end

            if @record.modified_at
              h.tag :small, class: "ml-1 text-muted" do
                h.tag :i, class: "fas fa-pencil-alt"
              end
            end
          end
        end

        h.tag :div, class: "col-auto pl-0" do
          if @current_user
            h.html Dropdowns::RecordOptionsDropdownComponent2.new(view_context,
              current_user: @current_user,
              record: @record
            ).render
          end
        end
      end
    end
  end
end
