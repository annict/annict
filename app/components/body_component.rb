# frozen_string_literal: true

class BodyComponent < ApplicationComponent
  def initialize(view_context, height: nil, format: :simple, class_name: "")
    super view_context
    @height = height
    @format = format
    @class_name = class_name
  end

  def render
    build_html do |h|
      h.tag :div,
        class: "c-body #{class_name}",
        data_controller: "body",
        data_body_height: height do
        h.tag :div, class: "c-body__content", data_body_target: "content" do
          yield h
        end

        h.tag :div,
          class: "c-body__read-more-background d-none w-100",
          data_body_target: "readMoreBackground"

        h.tag :div,
          class: "c-body__read-more-button d-none text-center w-100",
          data_body_target: "readMoreButton",
          data_action: "click->body#readMore" do
          h.tag :div, class: "c-body__read-more-content small u-fake-link w-100" do
            h.tag :i, class: "fal fa-chevron-double-down mr-1"
            h.text t("messages._components.body.view_full_text")
          end
        end
      end
    end
  end

  private

  attr_reader :class_name, :height

  def render_content
    case @format
    when :simple
      simple_format(content)
    when :markdown
      helpers.render_markdown(content)
    when :html
      content
    end
  end
end
