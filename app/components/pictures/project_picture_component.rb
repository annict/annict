# frozen_string_literal: true

module Pictures
  class ProjectPictureComponent < ApplicationComponent
    def initialize(project:, width:, alt: "", class_name: "")
      @project = project
      @width = width
      @height = @width
      @alt = alt.presence || project.name
      @class_name = class_name
    end

    private

    def source_srcset(format)
      [
        "#{helpers.ann_project_image_url(@project, width: @width, format: format)} 1x",
        "#{helpers.ann_project_image_url(@project, width: @width * 2, format: format)} 2x"
      ].join(", ")
    end
  end
end
