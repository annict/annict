module ItemDecoratorCommon
  extend ActiveSupport::Concern

  included do
    def to_values
      model.class::DIFF_FIELDS.inject({}) do |hash, field|
        hash[field] = case field
        when :url
          h.link_to(send(:url).truncate(30), send(:url), target: "_blank")
        when :tombo_image
          h.annict_image_tag(self, :tombo_image, size: "200x200")
        else
          send(field)
        end

        hash
      end
    end
  end
end
