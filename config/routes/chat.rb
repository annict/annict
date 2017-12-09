# frozen_string_literal: true

scope module: :chat do
  constraints(subdomain: "chat") do
    root "home#index", as: :chat_root
  end
end
