# frozen_string_literal: true

# Scroll to specific element helper
# https://stackoverflow.com/questions/17623075/auto-scroll-a-button-into-view-with-capybara-and-selenium

module ScrollToElement
  def scroll_to(selector)
    script = <<-JS
      // The element will be aligned to the center
      // https://developer.mozilla.org/en-US/docs/Web/API/Element/scrollIntoView
      arguments[0].scrollIntoView({ block: "center" });
    JS

    element = find(selector, visible: false)
    Capybara.current_session.driver.browser.execute_script(script, element.native)
  end
end

RSpec.configure do |config|
  config.include ScrollToElement, type: :system
end
