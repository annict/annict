require 'spec_helper'

describe 'トップページ' do
  let(:cover_work_id) { eval(ENV['ANNICT_COVER_IMAGE_DATA']).first['work_id'] }
  let!(:work) { create(:work, :with_item, id: cover_work_id) }

  context 'ログインしていないとき' do
    before do
      visit '/'
    end

    it 'アクセスするとページが表示される' do
      expect(page).to have_content('見たアニメを記録して、発見しよう')
    end

    it 'Twitterでログインをクリックするとユーザ登録ページが表示される' do
      find('.welcome').click_link('Twitterアカウントでログイン')

      expect(current_path).to eq '/users/sign_up'
    end

    it 'Facebookでログインをクリックするとユーザ登録ページが表示される' do
      find('.welcome').click_link('Facebookアカウントでログイン')

      expect(current_path).to eq '/users/sign_up'
    end
  end

  context 'ログインしているとき' do
    let(:user) { create(:registered_user) }

    before do
      login_as(user, scope: :user)
    end

    context 'アクティビティが存在しないとき' do
      before do
        visit '/'
      end

      it 'アクティビティが存在しない旨を表示すること' do
        expect(page).to have_content('アクティビティはありませんでした')
      end
    end

    context '自分がチェックインしているとき' do
      let(:work)    { create(:work, :with_item) }
      let(:episode) { create(:episode, work: work) }

      before do
        user.checkins.create(episode: episode, comment: 'おもしろかったよ')

        visit '/'
      end

      it 'アクティビティに自分のチェックイン情報が表示されること', js: true do
        expect(page).to have_content('おもしろかったよ')
      end
    end

    context 'フォローしている人がチェックインしているとき' do
      let(:following_user) { create(:registered_user) }
      let(:work)           { create(:work, :with_item) }
      let(:episode)        { create(:episode, work: work) }

      before do
        user.follow(following_user)
        user.checkins.create(episode: episode, comment: 'たのしかったよ')

        visit '/'
      end

      it 'アクティビティにフォローしている人のチェックイン情報が表示されること', js: true do
        expect(page).to have_content('たのしかったよ')
      end
    end
  end
end
