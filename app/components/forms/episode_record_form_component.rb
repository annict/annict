# frozen_string_literal: true

module Forms
  class EpisodeRecordFormComponent < ApplicationV6Component
    def initialize(view_context, form:, current_user:)
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
          controller: "episode-record-form",
          action: "turbo:submit-start->episode-record-form#handleSubmitStart turbo:submit-end->episode-record-form#handleSubmitEnd",
          "episode-record-form-target": "form"
        }
      ) do |f|
        build_html do |h|
          h.html V6::ErrorPanelComponent.new(view_context, stimulus_controller: "episode-record-form").render

          h.tag :div, class: "mb-2" do
            h.html ButtonGroups::RecordRatingButtonGroupComponent.new(view_context, form: f, rating_field: :rating).render
          end

          h.tag :div, class: "mb-3" do
            h.html V6::Textareas::RecordTextareaComponent.new(
              view_context,
              form: f,
              optional_textarea_classname: "form-control",
              textarea_name: "episode_record_form[comment]"
            ).render
          end

          h.tag :div, class: "row" do
            h.tag :div, class: "col" do
              if @current_user&.authorized_to?(:twitter, shareable: true)
                h.tag :div, class: "form-check" do
                  h.tag :label, class: "form-check-label" do
                    h.html f.check_box(:share_to_twitter, class: "form-check-input", checked: @current_user.share_record_to_twitter?)
                    h.text t("messages._common.share_to_twitter")
                  end
                end
              end
            end

            h.tag :div, class: "col" do
              h.tag :div, class: "text-center" do
                h.html f.submit((f.object.persisted? ? t("verb.update") : t("verb.track")), class: "btn btn-primary", data: {"episode-record-form-target": "submitButton"})
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
      @form.persisted? ? view_context.record_path(@current_user.username, @form.record.id) : view_context.episode_record_list_path(@form.episode.id)
    end
  end
end
