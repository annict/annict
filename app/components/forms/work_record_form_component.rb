# frozen_string_literal: true

module Forms
  class WorkRecordFormComponent < ApplicationV6Component
    def initialize(view_context, form:)
      super view_context
      @form = form
      @work = @form.work
      @record = @form.record
    end

    def render
      form_with(
        model: @form,
        url: form_url,
        method: form_method,
        data: {
          controller: "forms--work-record-form",
          action: "turbo:submit-start->forms--work-record-form#handleSubmitStart turbo:submit-end->forms--work-record-form#handleSubmitEnd",
          forms__work_record_form_target: "form"
        }
      ) do |f|
        build_html do |h|
          h.html ErrorPanelV6Component.new(view_context, stimulus_controller: "forms--work-record-form").render

          h.tag :div, class: "g-3 mb-3 row" do
            h.tag :div, class: "col-12 col-lg-6 col-xxl-4" do
              h.html f.label(:rating, class: "form-label small")
              h.html ButtonGroups::RecordRatingButtonGroupComponent.new(view_context, form: f, rating_field: :rating).render
            end

            h.tag :div, class: "col-12 col-lg-6 col-xxl-4" do
              h.html f.label(:animation_rating, class: "form-label small")
              h.html ButtonGroups::RecordRatingButtonGroupComponent.new(view_context, form: f, rating_field: :animation_rating).render
            end

            h.tag :div, class: "col-12 col-lg-6 col-xxl-4" do
              h.html f.label(:character_rating, class: "form-label small")
              h.html ButtonGroups::RecordRatingButtonGroupComponent.new(view_context, form: f, rating_field: :character_rating).render
            end

            h.tag :div, class: "col-12 col-lg-6 col-xxl-4" do
              h.html f.label(:story_rating, class: "form-label small")
              h.html ButtonGroups::RecordRatingButtonGroupComponent.new(view_context, form: f, rating_field: :story_rating).render
            end

            h.tag :div, class: "col-12 col-lg-6 col-xxl-4" do
              h.html f.label(:music_rating, class: "form-label small")
              h.html ButtonGroups::RecordRatingButtonGroupComponent.new(view_context, form: f, rating_field: :music_rating).render
            end
          end

          h.tag :div do
            h.html Textareas::RecordTextareaComponent.new(view_context,
              form: f, optional_textarea_classname: "form-control", textarea_name: "forms_work_record_form[body]"
            ).render
          end

          h.tag :div, class: "mb-3" do
            h.html Collapses::RecordFormOptionsCollapseComponent.new(
              view_context,
              form: f
            ).render
          end

          h.tag :div, class: "row" do
            h.tag :div, class: "col" do
              if current_user&.authorized_to?(:twitter, shareable: true)
                h.tag :div, class: "form-check" do
                  h.tag :label, class: "form-check-label" do
                    h.html f.check_box(:share_to_twitter, class: "form-check-input", checked: current_user.share_record_to_twitter?)
                    h.tag :i, class: "fab fa-twitter u-text-twitter"
                  end
                end
              end
            end

            h.tag :div, class: "col" do
              h.tag :div, class: "text-center" do
                h.html f.submit((f.object.persisted? ? t("verb.update") : t("verb.track")), {
                  class: "btn btn-primary",
                  data: {"forms--work-record-form-target": "submitButton"}
                })
              end
            end

            h.tag :div, class: "col"
          end
        end
      end
    end

    private

    def form_method
      @form.persisted? ? :patch : :post
    end

    def form_url
      @form.persisted? ? view_context.internal_api_record_path(@record.id) : view_context.internal_api_work_record_list_path(@work.id)
    end
  end
end
