# frozen_string_literal: true

module GraphQL
  module Define
    class DefinedObjectProxy
      include Imgix::Rails::UrlHelper
      include ImageHelper
    end
  end
end
