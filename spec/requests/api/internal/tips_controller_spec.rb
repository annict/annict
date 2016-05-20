# frozen_string_literal: true

describe Api::Internal::TipsController do
  let(:user) { create(:registered_user) }
  let(:tip) { create(:status_tip) }

  before do
    sign_in user
  end

  describe 'POST finish' do
    before do
      post :finish, partial_name: tip.partial_name
    end

    it '200が返ること' do
      expect(response.status).to eq(200)
    end

    it 'tipsが終了すること' do
      expect(user.finished_tips.count).to eq(1)
    end
  end
end
