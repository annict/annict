# frozen_string_literal: true

module Forms
  class RecordFormComponent < ApplicationV6Component
    def initialize(view_context, form:)
      super view_context
      @form = form
      @current_user = current_user
    end

    def render
      form_with(
        model: @form,
        url: form_url,
        method: form_method,
        data: {
          controller: "forms--record-form",
          action: "turbo:submit-start->forms--record-form#handleSubmitStart turbo:submit-end->forms--record-form#handleSubmitEnd",
          forms__record_form_target: "form"
        }
      ) do |f|
        build_html do |h|
          h.html ErrorPanelV6Component.new(view_context, stimulus_controller: "forms--record-form").render

          h.html f.hidden_field(:work_id)
          h.html f.hidden_field(:episode_id)

          h.tag :div, class: "mb-2" do
            h.html ButtonGroups::RecordRatingButtonGroupComponent.new(view_context, form: f, rating_field: :rating).render
          end

          h.tag :div, class: "mb-3" do
            h.html Textareas::RecordTextareaComponent.new(
              view_context,
              form: f,
              optional_textarea_classname: "form-control",
              textarea_name: "forms_record_form[body]"
            ).render
          end

          h.tag :div, class: "row" do
            h.tag :div, class: "col" do
              if @current_user&.authorized_to?(:twitter, shareable: true)
                h.tag :div, class: "form-check" do
                  h.tag :label, class: "form-check-label" do
                    h.html f.check_box(:share_to_twitter, class: "form-check-input", checked: @current_user.share_record_to_twitter?)
                    h.tag :i, class: "fab fa-twitter u-text-twitter"
                  end
                end
              end
            end

            h.tag :div, class: "col" do
              h.tag :div, class: "text-center" do
                h.html f.submit((f.object.persisted? ? t("verb.update") : t("verb.track")), class: "btn btn-primary", data: {"forms--record-form-target": "submitButton"})
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
      @form.persisted? ? view_context.internal_api_record_path(@current_user.username, @form.record.id) : view_context.internal_api_record_list_path
    end
  end
end
