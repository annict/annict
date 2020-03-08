# frozen_string_literal: true

module API
  module V1
    module Me
      class IndexShowParams
        include ActiveParameter

        param :fields

        validates :fields,
          allow_blank: true,
          fields_params: true
      end
    end
  end
end
