# frozen_string_literal: true

module V6
  class RecordHeaderComponent < V6::ApplicationComponent
    def initialize(view_context, record:)
      super view_context
      @record = record
      @user = @record.user
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-record-header row" do
          h.tag :div, class: "col-auto pe-0" do
            link_to view_context.profile_path(@record.user.username) do
              h.html V6::Pictures::AvatarPictureComponent.new(view_context,
                user: @user,
                width: 50,
                mb_width: 50).render
            end
          end

          h.tag :div, class: "col" do
            h.tag :div do
              h.tag :a, href: view_context.profile_path(@record.user.username), class: "fw-bold me-1 text-body" do
                h.tag :span, class: "me-1" do
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
                h.tag :small, class: "ms-1 text-muted" do
                  h.tag :i, class: "fas fa-pencil-alt"
                end
              end
            end
          end

          h.tag :div, class: "col-auto ps-0" do
            if current_user && RecordPolicy.new(current_user, @record).update?
              h.html V6::Dropdowns::RecordOptionsDropdownComponent.new(view_context, record: @record).render
            end
          end
        end
      end
    end
  end
end
