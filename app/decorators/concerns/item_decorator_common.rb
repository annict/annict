module ItemDecoratorCommon
  extend ActiveSupport::Concern

  included do
    def to_values
      model.class::DIFF_FIELDS.inject({}) do |hash, field|
        hash[field] = case field
        when :url
          h.link_to(send(:url).truncate(30), send(:url), target: "_blank")
        when :tombo_image
          image_url = h.tombo_thumb_url(self, :tombo_image, "w:200,h:200")
          h.image_tag(image_url, size: "200x200")
        when :main
          send(field) ? "使用する" : "使用しない"
        else
          send(field)
        end

        hash
      end
    end
  end
end
