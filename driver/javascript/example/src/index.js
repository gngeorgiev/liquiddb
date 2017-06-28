document.addEventListener(
    'DOMContentLoaded',
    function() {
        const field = document.getElementById('field');

        const pointSize = 20;
        const width = window.innerWidth - pointSize;
        const height = window.innerHeight - pointSize;

        const settings = {
            pointSize,
            gridWidth: width / pointSize,
            gridHeight: height / pointSize,
            frameInterval: 50,
            backgroundColor: '#f3e698'
        };
        // Create the game object. The settings object is NOT required.
        // The parentElement however is required
        const game = new SnakeJS(field, settings);
    },
    true
);
