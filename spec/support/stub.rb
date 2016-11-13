RSpec.configure do |config|
  config.before :suite do
    ApplicationHelper.module_eval do
      def ann_image_url(*)
        "#{ENV.fetch('ANNICT_URL')}/dummy_image"
      end
    end
  end
end
