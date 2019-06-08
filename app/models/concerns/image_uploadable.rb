# frozen_string_literal: true

module ImageUploadable
  extend ActiveSupport::Concern

  included do
    def uploaded_file(field, size: :master)
      send(field).instance_of?(Hash) ? send(field)[size] : send(field)
    end
  end
end
