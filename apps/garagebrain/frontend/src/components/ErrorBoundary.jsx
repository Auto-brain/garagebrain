import { Component } from 'react';

// ErrorBoundary ловит ошибки рендера в поддереве и показывает понятное
// сообщение вместо белого экрана. Без него любая исключительная ситуация
// в компоненте роняет всё приложение.
export default class ErrorBoundary extends Component {
  constructor(props) {
    super(props);
    this.state = { error: null };
  }

  static getDerivedStateFromError(error) {
    return { error };
  }

  componentDidCatch(error, info) {
    console.error('UI error:', error, info);
  }

  handleReset = () => {
    this.setState({ error: null });
    if (this.props.onReset) this.props.onReset();
  };

  render() {
    if (!this.state.error) return this.props.children;

    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 p-6">
        <div className="bg-white dark:bg-slate-800 rounded-2xl shadow-xl p-8 max-w-md text-center">
          <div className="text-4xl mb-3">⚠️</div>
          <h1 className="text-lg font-bold text-gray-800 mb-2">Что-то пошло не так</h1>
          <p className="text-gray-500 text-sm mb-5">
            Произошла ошибка в интерфейсе. Это не повредило ваши данные — попробуйте
            обновить страницу.
          </p>
          <div className="flex gap-3 justify-center">
            <button
              onClick={() => window.location.reload()}
              className="bg-blue-600 text-white px-5 py-2.5 rounded-lg font-medium hover:bg-blue-700 transition"
            >
              Обновить
            </button>
            <button
              onClick={this.handleReset}
              className="bg-gray-100 text-gray-700 px-5 py-2.5 rounded-lg font-medium hover:bg-gray-200 transition"
            >
              Попробовать снова
            </button>
          </div>
          {this.state.error?.message && (
            <p className="text-xs text-gray-400 mt-4 break-words">{this.state.error.message}</p>
          )}
        </div>
      </div>
    );
  }
}
